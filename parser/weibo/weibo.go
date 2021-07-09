// Package weibo 下载、发送微博指定用户的所有图集到 PicTG
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
	"math/rand"
	"sort"
	"strings"
	"time"
)

const (
	// API
	mymblogAPI  = "https://weibo.com/ajax/statuses/mymblog?uid=%s&page=%d&feature=1"
	downloadAPI = "https://weibo.com/ajax/common/download?pid=%s"
)

// DownloadAll 下载微博指定用户的所有图集
func DownloadAll(target *conf.Target) {
	// 请求头
	headers := map[string]string{
		"Cookie":  target.Cookie,
		"Referer": "https://weibo.com",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
	}

	// 保存的指定用户所有图集的下载信息的文件名
	var postsName string
	// 上次程序运行完时已保存到的帖子 ID
	idDone := target.IDDone

	// 先下载之前失败的图集
	// 失败的任务不需要先排序，因为全部需要下载，不需要判断结束
	retryTasks := logger.GetFailLog()
	// 计数
	logger.Info.Printf("[%s][%s] 重试下载 %d 个图集\n", target.Plat, target.ID, len(retryTasks))
	for _, task := range retryTasks {
		// 重试时设置标志为重试，以在下载成功后删除失败记录中该图集的记录
		task.IsRetry = true
		task.Header = headers
		task.IDDonePtr = &target.IDDone
		logger.Info.Printf("[%s][%s] 重试下载图集'%s'\n", target.Plat, target.ID, task.ID)
		worker.TasksCh <- task
	}

	logger.Info.Printf("[%s][%s] 已重发上次失败的图集，将开始继续下载新图集\n", target.Plat, target.ID)

	// 该用户的所有图集
	var allAlbum = make(map[string]comm.Album)

	// 判断图集的下载信息的文件是否存在，决定是联网获取，还是从本地读取
	// 本地文件命名为："平台名_ID.json"，如"weibo_6032474791.json"
	postsName = fmt.Sprintf("%s_%s.json", comm.TagWeibo, target.ID)
	exists, err := dofile.Exists(postsName)
	logger.Fatal("判断图集数据的文件是否存在时出错", err)
	if exists {
		allAlbum = unMarshalPosts(postsName)
		logger.Info.Printf("[%s][%s] 已从文件读取该用户的图集数据\n", comm.TagWeibo, target.ID)
	} else {
		logger.Info.Printf("[%s][%s] 没有该用户的图集数据的文件\n", comm.TagWeibo, target.ID)
		return
		/*
			allAlbum = getAllAlbum(target.ID, idDone, &headers)
			marshalPosts(allAlbum, postsName)
			logger.Info.Printf("[%s]已保存用户'%s'的所有图集到文件'%s'\n", comm.TagWeibo, target.ID, postsName)
		*/
	}

	// 排序
	idstrList := make([]string, len(allAlbum))
	i := 0
	for idstr := range allAlbum {
		idstrList[i] = idstr
		i++
	}
	sort.Strings(idstrList)

	// 根据帖子的 idstr 从小打到升序发送图集
	// 这样做方便保存发送的进度，可用于下次运行程序时，读取进度后继续发送
	// 当完成此次发送任务后，idstr 将为最新（即序号最大）
	// 这更加方便了获取更新，即在获取更新帖时，依然从第一页遍历，当读到该 idstr 时为止（不应包括 idstr）

	// 上次运行结束后（完成或中断），已完成到的索引
	lastIndex := sort.SearchStrings(idstrList, idDone)
	// 因为 SearchStrings 当没有找到时会返回新元素应该插入的索引，需要矫正特殊情况下的索引
	if idDone == "" || lastIndex == len(idstrList) {
		lastIndex = -1
	}
	// 继续任务
	taskIdstrList := idstrList[lastIndex+1:]
	// 计数
	logger.Info.Printf("[%s][%s] 新添加下载 %d 个图集\n", target.Plat, target.ID, len(taskIdstrList))

	for _, idstr := range taskIdstrList {
		// 因为 ID 按从小到大的增序排序，跳过已保存过的 ID
		if idDone != "" && strings.Compare(idstr, idDone) <= 0 {
			continue
		}
		// 跳过没有图集链接的任务
		if len(allAlbum[idstr].URLs) == 0 || len(allAlbum[idstr].URLsM) == 0 {
			continue
		}
		// logger.Info.Printf("准备发送图集'%s'\n", idstr)
		worker.TasksCh <- comm.Album{Tag: comm.TagWeibo, ID: idstr, Created: allAlbum[idstr].Created,
			URLs: allAlbum[idstr].URLs, URLsM: allAlbum[idstr].URLsM, IDDonePtr: &target.IDDone, Header: headers}
	}

	// 任务完成
	logger.Info.Printf("[%s][%s] 已提交所有任务\n", target.Plat, target.ID)
}

// 从文件解析图集信息
func unMarshalPosts(path string) map[string]comm.Album {
	bs, err := dofile.Read(path)
	logger.Fatal("读取图集数据的文件时出错", err)
	var posts map[string]comm.Album
	err = json.Unmarshal(bs, &posts)
	logger.Fatal("解析图集数据时出错", err)
	return posts
}

// 保存图集信息到文件
func marshalPosts(posts map[string]comm.Album, path string) {
	bs, err := json.MarshalIndent(posts, "", "  ")
	logger.Fatal("文本化图集数据时出错", err)
	_, err = dofile.Write(bs, path, dofile.OTrunc, 0644)
	logger.Fatal("将图集数据保存到文件时出错", err)
}

// 获取微博指定用户所有帖子的图集
func getAllAlbum(uid string, idDone string, headers *map[string]string) map[string]comm.Album {
	// 用于保存所有图集，将返回
	posts := make(map[string]comm.Album)

	// 读取所有帖子
	page := 1
	var postPage PostPage
	for {
		// 读取 API，解析
		bs, err := client.Client.Get(fmt.Sprintf(mymblogAPI, uid, page), *headers)
		logger.Fatal("联网获取图集数据时出错", err)
		err = json.Unmarshal(bs, &postPage)
		logger.Fatal("解析获取到底图集数据时出错", err)

		// 返回内容的帖子数量为 0 时，表示遍历完成，退出循环
		if len(postPage.Data.List) == 0 {
			return posts
		}

		// 遍历帖子
		for _, post := range postPage.Data.List {
			// 当读取的帖子的 idstr 和已保存的进度记录相同时，说明已完成任务，直接返回数据
			if post.Idstr == idDone {
				return posts
			}

			// 读取、保存该贴的图集
			task := comm.Album{
				Tag:     comm.TagWeibo,
				Caption: post.TextRaw,
				ID:      post.Idstr,
				Created: post.Created,
				URLs:    make([]string, len(post.PicIds)),
				URLsM:   nil,
				Header:  *headers,
			}
			for _, pid := range post.PicIds {
				task.URLs = append(task.URLs, fmt.Sprintf(downloadAPI, pid))
			}

			// 添加到所有图集中，以返回
			posts[post.Idstr] = task
		}
		logger.Info.Printf("[%s][%s] 已添加第 %d 页的图集\n", page)
		page++

		// 等待不固定的时间，以防被禁止访问
		r := rand.Intn(5)
		time.Sleep(time.Duration(r) * time.Second)
	}
}
