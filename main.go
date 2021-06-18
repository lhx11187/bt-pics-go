package main

import (
	"bt-pics-go/conf"
	"bt-pics-go/logger"
	"bt-pics-go/parser/weibo"
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
		conf.SaveWBIDStr()
		os.Exit(0)
	}()

	// 保存进度
	defer func() {
		if err := recover(); err != nil {
			conf.SaveWBIDStr()
		}
	}()

	// 执行任务
	weibo.DownloadAll("6032474791")
}
