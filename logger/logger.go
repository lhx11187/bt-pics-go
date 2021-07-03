package logger

import (
	"bt-pics-go/conf"
	"github.com/donething/utils-go/dolog"
	"log"
)

var (
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

const LogName = "run.log"

func init() {
	Info, Warn, Error = dolog.InitLog(LogName, dolog.DefaultFormat)
}

// SaveWhenExit 当崩溃时保存记录
func SaveWhenExit() {
	// 保存微博的进度
	conf.Save()

	// 保存失败的记录到文件
	logFailToFile()
}

func Fatal(err error) {
	if err != nil {
		panic(err)
	}
}
