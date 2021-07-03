package logger

import (
	"bt-pics-go/comm"
	"encoding/json"
	"github.com/donething/utils-go/dofile"
	"sync"
)

// 记录执行下载、发送任务失败的操作
const failLogName = "fail.log"

var (
	// 失败记录
	failLog = make(map[string]comm.Album)
	mu      sync.Mutex
)

func init() {
	// 读取出错记录
	exists, err := dofile.Exists(failLogName)
	Fatal(err)
	if exists {
		bs, err := dofile.Read(failLogName)
		Fatal(err)

		err = json.Unmarshal(bs, &failLog)
		Fatal(err)
	}
}

// GetFailLog 获取失败的记录
func GetFailLog() map[string]comm.Album {
	return failLog
}

// LogFail 记录出错
func LogFail(album comm.Album) {
	// 不需要记录请求头
	album.Header = nil
	album.IDDonePtr = nil
	mu.Lock()
	failLog[album.ID] = album
	mu.Unlock()
}

// LogRmFail 删除已重试成功的失败记录
func LogRmFail(album comm.Album) {
	mu.Lock()
	delete(failLog, album.ID)
	mu.Unlock()
}

// logFailToFile 保存失败记录到文件
func logFailToFile() {
	mu.Lock()
	if len(failLog) == 0 {
		mu.Unlock()
		return
	}
	bs, err := json.MarshalIndent(failLog, "", "  ")
	mu.Unlock()

	Fatal(err)
	_, err = dofile.Write(bs, failLogName, dofile.OTrunc, 0644)
	Fatal(err)
}
