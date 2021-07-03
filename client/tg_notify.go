package client

import (
	"bt-pics-go/conf"
	"bt-pics-go/logger"
	"github.com/donething/utils-go/dotgpush"
)

var (
	// TG 通知机器人的 token 和 chat id
	tg         *dotgpush.TGBot
	noTokenTG  string
	noChatIDTG string
)

func init() {
	conf.Mu.Lock()
	noTokenTG = conf.Conf.TG.NoToken
	noChatIDTG = conf.Conf.TG.NoChatID
	conf.Mu.Unlock()
	if noTokenTG != "" && noChatIDTG != "" {
		tg = dotgpush.NewTGBot(conf.Conf.TG.NoToken)
	}
}

// Notify 发送 TG 通知
func Notify(msg string) {
	if tg == nil {
		return
	}
	_, err := tg.SendMessage(noChatIDTG, msg)
	logger.Fatal(err)
}
