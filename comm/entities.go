package comm

// Task 向 worker 发送的任务
type Task struct {
	Tag    string             // 网站
	ID     string             // 任务的 ID
	Data   interface{}        // 发送的数据
	Header *map[string]string // 下载数据的请求头，可空
}
