package main

import (
	"bt-pics-go/client"
	"bt-pics-go/comm"
	"bt-pics-go/conf"
	"bt-pics-go/logger"
	"bt-pics-go/parser/weibo"
	"bt-pics-go/worker"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Go signal notification works by sending `os.Signal`
	// values on a channel. We'll create a channel to
	// receive these notifications (we'll also make one to
	// notify us when the program can exit).
	sigs := make(chan os.Signal, 2)
	// `signal.Notify` registers the given channel to
	// receive notifications of the specified signals.
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// This goroutine executes a blocking receive for
	// signals. When it gets one it'll print it out
	// and then notify the program that it can finish.
	go func() {
		<-sigs
		logger.Info.Printf("已收到中断信号，将立即保存进度后退出程序\n")
		logger.SaveWhenExit()
		os.Exit(0)
	}()

	// 当程序崩溃时保存进度
	defer func() {
		if err := recover(); err != nil {
			logger.Error.Printf("程序崩溃：%s\n", err)
			logger.SaveWhenExit()
		}
	}()

	// 初始化下载、发送任务
	workerCount := conf.Conf.WorkerCount
	worker.InitWorker(workerCount)
	worker.WG.Add(workerCount)

	// 执行任务
	for i, target := range conf.Conf.PicTargets {
		switch target.Plat {
		case comm.TagWeibo:
			weibo.DownloadAll(&conf.Conf.PicTargets[i])
		}
	}

	// 等待任务完成
	// 注意：当整个过程都没有任务时，程序会 在 worker 一直等待任务，而发生阻塞，需要手动停止程序
	logger.Info.Println("正在等待任务完成…")
	worker.WG.Wait()

	// 已完成所有任务，准备结束程序
	logger.SaveWhenExit()
	client.Notify(fmt.Sprintf("[BT_PICS] [%s] 已完成任务", comm.TagWeibo))
}
