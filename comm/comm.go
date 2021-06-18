// Package comm 公共数据，不引用本工程下自己编写的库
package comm

const (
	// TagWeibo 微博
	TagWeibo = "weibo"

	// PassToLocal PassToTG PassToYike 在 worker 中对说下载数据的处理操作
	PassToLocal = "SAVE_LOCAL" // 保存到本地
	PassToTG    = "ToTG"       // 发送到 Telegram
	PassToYike  = "ToYike"     // 发送到一刻相册

	// Root 当 processor 的值为 PassToLocal 时，保存文件到的本地目录
	Root = "D:/Temp/Pics"
)
