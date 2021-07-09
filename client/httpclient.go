// Package client 执行 HTTP 请求
package client

import (
	"bt-pics-go/conf"
	"bt-pics-go/logger"
	"github.com/donething/utils-go/dohttp"
	"time"
)

var (
	// Client 执行 HTTP 请求的客户端
	Client = dohttp.New(30*time.Second, false, false)
)

func init() {
	// 如果配置中指定了代理，需要设置
	proxy := conf.Conf.Proxy
	if proxy != "" {
		err := Client.SetProxy(proxy)
		if err != nil {
			logger.Fatal("设置HTTP代理时出错", err)
		}
	}
}
