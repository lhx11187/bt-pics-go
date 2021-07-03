package conf

import (
	"encoding/json"
	"fmt"
	"github.com/donething/utils-go/dofile"
	"os"
	"path"
	"sync"
)

const (
	// Name 配置文件的名字
	Name = "bt-pics-go.json"

	// HandlerToLocal HandlerToTG HandlerToYike 在 worker 中对数据的处理方法
	HandlerToLocal = "ToLocal" // 保存到本地
	HandlerToTG    = "ToTG"    // 发送到 Telegram
	HandlerToYike  = "ToYike"  // 发送到一刻相册
)

var (
	// Conf 配置的实例
	Conf Config
	Mu   sync.Mutex

	// 配置文件所在的路径
	confPath string
)

type Target struct {
	// 上次发送到的帖子的 ID，用于从此 ID 开始继续发送
	ID     string `json:"id"`
	Plat   string `json:"plat"`
	IDDone string `json:"id_done"`
	Cookie string `json:"cookie"`
	Auth   string `json:"auth"`
}

type Config struct {
	// 工作池的容量
	WorkerCount int `json:"worker_count"`

	// 对文件数据的处理，可从常量中选择 Handler***
	Handler string `json:"handler"`

	// 当 Handler 的值为 HandlerToLocal 时，保存文件到的本地目录
	LocalRoot string `json:"local_root"`

	// 使用代理，为空表示不使用代理
	Proxy string `json:"proxy"`

	// 抓取的目标
	PicTargets []Target `json:"pic_targets"`

	// 推送
	// 一刻相册
	Yike struct {
		Bdstoken string `json:"bdstoken"`
		Cookie   string `json:"cookie"`
	} `json:"yike"`

	// Telegram 推送消息
	TG struct {
		NoToken       string `json:"no_token"`
		NoChatID      string `json:"no_chat_id"`
		PicSaveToken  string `json:"pic_save_token"`
		PicSaveChatID string `json:"pic_save_chat_id"`
	} `json:"tg"`
}

func init() {
	confPath = path.Join(Name)
	// 创建默认配置文件
	saveFile(confPath + ".bak")

	// 创建读取配置文件
	exist, err := dofile.Exists(confPath)
	fatal(err)
	if exist {
		fmt.Printf("读取配置文件：%s\n", confPath)
		bs, err := dofile.Read(confPath)
		fatal(err)
		err = json.Unmarshal(bs, &Conf)
		fatal(err)
	} else {
		fmt.Printf("创建配置文件：%s\n", confPath)
		Conf.PicTargets = append(Conf.PicTargets, Target{})
		bs, err := json.MarshalIndent(Conf, "", "  ")
		fatal(err)
		_, err = dofile.Write(bs, confPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		fatal(err)
	}
}

// Save 保存配置
func Save() {
	saveFile(confPath)
}

// 保存配置到文件
func saveFile(path string) {
	bs, err := json.MarshalIndent(Conf, "", "  ")
	fatal(err)
	_, err = dofile.Write(bs, path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	fatal(err)
}

// fatal 出错时，强制关闭程序
func fatal(err error) {
	if err != nil {
		panic(fmt.Errorf("处理配置文件出错：%w", err))
	}
}
