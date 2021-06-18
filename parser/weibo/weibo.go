// Package weibo 下载、发送微博指定用户的所有图集到 TG
package weibo

import (
	"bt-pics-go/client"
	"bt-pics-go/comm"
	"bt-pics-go/conf"
	"bt-pics-go/logger"
	"bt-pics-go/worker"
	"encoding/json"
	"fmt"
	"github.com/donething/utils-go/dofile"
	"github.com/donething/utils-go/dotgpush"
	"math/rand"
	"sort"
	"strings"
	"time"
)

const (
	mymblogAPI  = "https://weibo.com/ajax/statuses/mymblog?uid=%s&page=%d&feature=1"
	downloadAPI = "https://weibo.com/ajax/common/download?pid=%s"
)

var (
	headers map[string]string
)

func init() {
	headers = map[string]string{
		"Cookie":  conf.Conf.Weibo.Cookie,
		"Referer": "https://weibo.com",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
	}
}

// DownloadAll 下载微博指定用户的所有图集
func DownloadAll(uid string) {
	// 该用户的所有图集
	var allAlbum map[string][]dotgpush.Media
	// 保存的指定用户所有图集的下载信息的文件名
	var postsName string
	// 上次已保存到的帖子 ID
	lastIDStr := conf.Conf.Weibo.LastIDStr

	// 判断图集的下载信息的文件是否存在，决定是联网获取，还是从本地读取
	postsName = fmt.Sprintf("%s_%s.json", comm.TagWeibo, uid)
	exists, err := dofile.Exists(postsName)
	logger.Fatal(err)
	if exists {
		allAlbum = unMarshalPosts(postsName)
		logger.Info.Printf("[%s]已从文件读取用户'%s'的所有图集\n", comm.TagWeibo, uid)
	} else {
		allAlbum = getAllAlbum(uid, lastIDStr)
		marshalPosts(allAlbum, postsName)
		logger.Info.Printf("[%s]已保存用户'%s'的所有图集到文件'%s'\n", comm.TagWeibo, uid, postsName)
	}

	// 初始化下载、发送任务
	tasksCh, wg := worker.InitWorker(10)

	// 排序
	idList := make([]string, len(allAlbum))
	i := 0
	for idstr := range allAlbum {
		idList[i] = idstr
		i++
	}
	sort.Strings(idList)

	// 根据 帖子的 idstr 升序发送图集
	// 这样做方便保存发送的进度，可用于下次运行程序时，读取进度后继续发送
	// 当发送已有的图集后，idstr 将为最新（即序号最大）
	// 这更加方便了获取更新，即在获取更新帖时，依然从第一页遍历，当读到该 idstr 时为止（不应包括 idstr）
	for _, idstr := range idList {
		// 因为 ID 按从小到大的顺序排序，跳过已保存过的 ID
		if strings.Compare(idstr, lastIDStr) <= 0 {
			continue
		}
		logger.Info.Printf("准备发送图集'%s'\n", idstr)
		tasksCh <- comm.Task{Tag: comm.TagWeibo, ID: idstr, Data: allAlbum[idstr], Header: &headers}
	}

	// 等待下载、发送的任务完成
	wg.Wait()
	logger.Info.Printf("[%s]已完成所有下载、发送的任务\n", comm.TagWeibo)
	conf.SaveWBIDStr()
	_, err = client.TG.SendMessage(client.ChatID, "[微博]已完成所有下载、发送的任务")
	logger.Fatal(err)
}

// 获取微博指定用户所有帖子的图集
func getAllAlbum(uid string, lastIDStr string) map[string][]dotgpush.Media {
	// 用于保存所有图集，将返回
	posts := make(map[string][]dotgpush.Media)

	// 读取所有帖子
	page := 1
	var postPage PostPage
	for {
		// 读取 API，解析
		bs, err := client.Client.Get(fmt.Sprintf(mymblogAPI, uid, page), headers)
		logger.Fatal(err)
		err = json.Unmarshal(bs, &postPage)
		logger.Fatal(err)

		// 返回内容的帖子数量为 0 时，表示遍历完成，退出循环
		if len(postPage.Data.List) == 0 {
			return posts
		}

		// 遍历帖子
		for _, post := range postPage.Data.List {
			// 当读取的帖子的 idstr 和已保存的进度记录相同时，说明已完成任务，直接返回数据
			if post.Idstr == lastIDStr {
				return posts
			}

			/*
				// 记录此次下载任务的第一个帖子的 ID，作为进度保存
				if page == 1 && i == 0 {
					thisFirstIDStr = post.Idstr
				}
			*/

			// 读取、保存该贴的图集
			album := make([]dotgpush.Media, 0)
			for _, pid := range post.PicIds {
				// 跳过为空的下载链接
				// 可能图片因为敏感而已被删除
				if strings.TrimSpace(pid) == "" {
					continue
				}
				album = append(album, dotgpush.Media{
					Type:    dotgpush.Photo,
					Media:   fmt.Sprintf(downloadAPI, pid),
					Caption: "",
				})
			}
			// 设置该图集的标题
			if len(album) >= 1 {
				album[0].Caption = post.TextRaw
			}

			// 添加到所有图集中，以返回
			posts[post.Idstr] = album
		}
		logger.Info.Printf("[微博] 已添加第 %d 页的图集\n", page)
		page++

		// 等待不固定的时间，以防被禁止访问
		r := rand.Intn(5)
		time.Sleep(time.Duration(r) * time.Second)
	}
}

// 保存图集信息到文件
func marshalPosts(posts map[string][]dotgpush.Media, path string) {
	bs, err := json.MarshalIndent(posts, "", "  ")
	logger.Fatal(err)
	_, err = dofile.Write(bs, path, dofile.OTrunc, 0644)
	logger.Fatal(err)
}

// 从文件解析图集信息
func unMarshalPosts(path string) map[string][]dotgpush.Media {
	bs, err := dofile.Read(path)
	logger.Fatal(err)
	var posts map[string][]dotgpush.Media
	err = json.Unmarshal(bs, &posts)
	logger.Fatal(err)
	return posts
}
