package client

import (
	"bt-pics-go/conf"
	"github.com/donething/utils-go/dotgpush"
)

var (
	// TG 发送消息
	TG     = dotgpush.NewTGBot(conf.Conf.TG.PicSaveToken)
	ChatID = conf.Conf.TG.PicSaveChatID
)
