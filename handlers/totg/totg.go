package totg

import (
	"bt-pics-go/client"
	"bt-pics-go/comm"
	"bt-pics-go/conf"
	"fmt"
	"github.com/donething/utils-go/dotgpush"
)

var (
	// Telegram 发送消息
	tg     = dotgpush.NewTGBot(conf.Conf.TG.PicSaveToken)
	chatID = conf.Conf.TG.PicSaveChatID
)

// Send 发送到 PicTG
func Send(album comm.Album) error {
	// 下载图集
	medias := make([]dotgpush.Media, len(album.URLsM))
	for i := 0; i < len(album.URLsM); i++ {
		// 将下载链接转为对应文件的二进制数组数据
		bs, err := client.Client.Get(album.URLsM[i], album.Header)
		if err != nil {
			return fmt.Errorf("下载文件'%s'出错：%s\n", album.URLs[i], err)
		}
		medias[i] = dotgpush.Media{
			Type:    dotgpush.Photo,
			Media:   bs,
			Caption: "",
		}
	}

	// 设置图集的标题
	if len(medias) > 0 {
		medias[0].Caption = album.Caption
	}

	// 发送图集
	msg, err := tg.SendMediaGroup(chatID, medias)
	if err != nil {
		return fmt.Errorf("发送文件'%s'出错，将保存信息后，跳到下个任务：%s\n", album.ID, err)
	}
	if msg == nil || !msg.Ok {
		return fmt.Errorf("发送文件'%s'失败，将保存信息后，跳到下个任务：%s\n", album.ID, msg.Description)
	}
	return nil
}
