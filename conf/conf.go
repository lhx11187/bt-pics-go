package conf

import (
	"encoding/json"
	"fmt"
	"github.com/donething/utils-go/dofile"
	"os"
	"path"
	"sync"
)

type Config struct {
	// 对文件数据的处理，可选择 "SAVE_LOCAL"：保存到本地，"ToTG"：发送到 TG ，"ToYike"：发送到一刻相册
	Processor string `json:"processor"`

	// 使用代理，为空表示不使用代理
	Proxy string `json:"proxy"`

	// Telegram 推送消息
	TG struct {
		PicSaveToken  string `json:"pic_save_token"`
		PicSaveChatID string `json:"pic_save_chat_id"`
	} `json:"tg"`

	// 微博
	Weibo struct {
		// 上次发送到的帖子的 ID，用于从此 ID 开始继续发送
		LastIDStr string `json:"last_id_str"`
		Cookie    string `json:"cookie"`
	} `json:"weibo"`

	// 一刻相册
	Yike struct {
		Bdstoken string `json:"bdstoken"`
		Cookie   string `json:"cookie"`
	} `json:"yike"`
}

const (
	// Name 配置文件的名字
	Name = "bt-pics-go.json"
)

var (
	// Conf 配置的实例
	Conf Config

	// LastIDStrTmp 上次已保存到的帖子 ID
	LastIDStrTmp string
	IDMu         sync.Mutex

	// 配置文件所在的路径
	confPath string
)

func init() {
	confPath = path.Join(Name)
	exist, err := dofile.Exists(confPath)
	fatal(err)
	if exist {
		fmt.Printf("读取配置文件：%s\n", confPath)
		bs, err := dofile.Read(confPath)
		fatal(err)
		err = json.Unmarshal(bs, &Conf)
		fatal(err)
		// 保存进度到临时变量
		IDMu.Lock()
		LastIDStrTmp = Conf.Weibo.LastIDStr
		IDMu.Unlock()
	} else {
		fmt.Printf("创建配置文件：%s\n", confPath)
		bs, err := json.MarshalIndent(Conf, "", "  ")
		fatal(err)
		_, err = dofile.Write(bs, confPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		fatal(err)
	}
}

// Save 保存配置
func Save() {
	bs, err := json.MarshalIndent(Conf, "", "  ")
	fatal(err)
	_, err = dofile.Write(bs, confPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	fatal(err)
}

// SaveWBIDStr 保存微博的进度
func SaveWBIDStr() {
	IDMu.Lock()
	Conf.Weibo.LastIDStr = LastIDStrTmp
	Save()
	IDMu.Unlock()
}

// fatal 出错时，强制关闭程序
func fatal(err error) {
	if err != nil {
		panic(fmt.Errorf("处理配置文件出错：%w", err))
	}
}
