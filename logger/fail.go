package logger

import (
	"encoding/json"
	"github.com/donething/utils-go/dofile"
	"sync"
	"bt-pics-go/comm"
)

// 记录执行下载、发送任务失败的操作
const failLogName = "fail.log"

var (
	failLog = make([]comm.Task, 0)
	mu      sync.Mutex
)

// LogFail 记录出错
func LogFail(task comm.Task) {
	// 不需要记录请求头
	task.Header = nil
	mu.Lock()
	failLog = append(failLog, task)
	bs, err := json.MarshalIndent(failLog, "", "  ")
	mu.Unlock()
	Fatal(err)

	_, err = dofile.Write(bs, failLogName, dofile.OAppend, 0644)
	Fatal(err)
}

// ReadFail 读取出错记录
func ReadFail() []comm.Task {
	var logs []comm.Task
	bs, err := dofile.Read(failLogName)
	Fatal(err)

	err = json.Unmarshal(bs, &logs)
	Fatal(err)

	return logs
}
