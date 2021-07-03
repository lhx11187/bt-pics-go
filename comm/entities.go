package comm

// Album 向 worker 发送的专辑
type Album struct {
	// 基础，必需
	Tag     string   `json:"tag"`              // 网站
	Caption string   `json:"caption"`          // 标题
	Created int64    `json:"created"`          // 创建时间
	ID      string   `json:"id"`               // 任务的 ID
	URLs    []string `json:"urls"`             // 发送的数据，如 URL 的数组（若为图片则为最大分辨率）
	URLsM   []string `json:"urls_m,omitempty"` // 若为图片，则为中等分辨率

	// 后续设置，可空
	IDDonePtr *string           `json:"id_done_ptr"`        // 已成功完成到的进度的指针(即配置文件中 ID 的指针)
	Header    map[string]string `json:"header,omitempty"`   // 下载文件的请求头，可空
	IsRetry   bool              `json:"is_retry,omitempty"` // 该任务是否为重试（重试成功则要删除失败记录）
}
