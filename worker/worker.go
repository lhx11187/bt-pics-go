package worker

import (
	"bt-pics-go/client"
	"bt-pics-go/comm"
	"bt-pics-go/conf"
	"bt-pics-go/logger"
	"bt-pics-go/yike"
	"fmt"
	"github.com/donething/utils-go/dofile"
	"github.com/donething/utils-go/dotgpush"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	// 对文件数据的处理

	// Telegram 的 token 信息
	tg     *dotgpush.TGBot
	chatID string
	// 网络链接的正则
	urlReg = regexp.MustCompile(`^(https?|ftps?)://`)
)

// InitWorker 初始化工作协程
func InitWorker(workerCount int) (chan comm.Task, *sync.WaitGroup) {
	// 设置处理器
	// 是否需要设置 TG 推送信息
	if conf.Conf.Processor == comm.PassToTG {
		tg = dotgpush.NewTGBot(conf.Conf.TG.PicSaveToken)
		chatID = conf.Conf.TG.PicSaveChatID
	}

	// 创建一个有缓冲的通道来管理工作
	tasksCh := make(chan comm.Task, workerCount)

	// 等待所有工人工作完成，即可退出
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// 启动goroutine来完成工作
	for id := 1; id <= workerCount; id++ {
		go func(gr int) {
			defer wg.Done()
			worker(tasksCh, &wg, gr)
		}(id)
	}
	logger.Info.Println("工作 goroutine 已准备就绪")
	return tasksCh, &wg
}

// 工作
func worker(tasksCh chan comm.Task, wg *sync.WaitGroup, id int) {
	defer wg.Done()
	for {
		// 等待分配工作
		task, ok := <-tasksCh
		if !ok {
			// 这意味着通道已经空了，并且已被关闭
			fmt.Printf("Worker: %d : Shutting Down\n", id)
			return
		}

		// 显示我们开始工作了
		// fmt.Printf("Worker: %d : Started %s\n", id, task)

		// 根据标签匹配 下载、发送媒体文件 的操作
	Next:
		switch task.Tag {
		case comm.TagWeibo:
			// 将数据转为 媒体集
			data := task.Data.([]dotgpush.Media)

			// 下载前的处理
			var destDir string
			if conf.Conf.Processor == comm.PassToLocal {
				// 保存到本地文件时，先创建目录
				destDir = filepath.Join(comm.Root, task.ID)
				err := os.MkdirAll(destDir, 0644)
				if err != nil {
					logger.Error.Printf("创建目录'%s'出错：%s\n", destDir, err)
					logger.LogFail(task)
					break Next
				}
			}

			// 下载二进制文件
			for i, media := range data {
				// media 文件为 URL 时，先下载，再发送到 TG
				if url, ok := media.Media.(string); ok && urlReg.MatchString(url) {
					// 将下载链接转为对应文件的二进制数组数据
					bs, err := client.Client.Get(url, *task.Header)
					if err != nil {
						logger.Error.Printf("下载媒体文件'%s'出错，将保存该图集的信息后，跳到下个任务：%s\n",
							url, err)
						logger.LogFail(task)
						break Next
					}

					// 提取文件名
					filename := filepath.Join(destDir, filepath.Base(url))
					if i := strings.Index(filename, "?"); i >= 0 {
						filename = filename[:i]
					}

					// 发送到 TG
					if conf.Conf.Processor == comm.PassToTG {
						// 使用索引访问 data 才能生效，对 media.Media 的修改不会作用到 data
						data[i].Media = bs
						data[i].Type = dotgpush.Photo
					} else if conf.Conf.Processor == comm.PassToLocal {
						_, err = dofile.Write(bs, filename, dofile.OTrunc, 0644)
						if err != nil {
							logger.Error.Printf("将数据保存到文件'%s'时出错：%s\n", filename, err)
							logger.LogFail(task)
							break Next
						}
					} else if conf.Conf.Processor == comm.PassToYike {
						// 发送到一刻相册
						yk := yike.New(bs, fmt.Sprintf("/%s/%s", task.ID, filename))
						err = yk.UploadFile()
						if err != nil {
							logger.Error.Printf("发送文件(%dKB)到一刻相册时出错：%s\n", len(bs)/1024, err)
							logger.LogFail(task)
							break Next
						}
					}
				}
			}

			// 发送到 TG
			if conf.Conf.Processor == comm.PassToTG {
				// 发送图集
				msg, err := tg.SendMediaGroup(chatID, data)
				if err != nil {
					logger.Error.Printf("发送文件出错，将保存该图集'%s'的信息后，跳到下个任务：%s\n", task.ID, err)
					logger.LogFail(task)
					break Next
				}
				if msg == nil || !msg.Ok {
					logger.Error.Printf("发送文件失败，将保存该图集'%s'的信息后，跳到下个任务：%s\n",
						task.ID, msg.Description)
					logger.LogFail(task)
					break Next
				}
			}

			// 仅当完成本次下载、发送任务时，才保存进度
			conf.IDMu.Lock()
			if strings.Compare(task.ID, conf.LastIDStrTmp) > 0 {
				conf.Conf.Weibo.LastIDStr = task.ID
			}
			conf.IDMu.Unlock()

			logger.Info.Printf("[Worker%d][%s]已完成发送图集'%s'\n", id, task.Tag, task.ID)
		}

		// 显示我们完成了工作
		// fmt.Printf("Worker: %d : Completed %s\n", id, task)
	}
}
