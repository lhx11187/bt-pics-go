package weibo

// PostPage 一次请求某页 API 时的返回信息
type PostPage struct {
	Data struct {
		List []struct {
			Idstr   string   `json:"idstr"`
			PicIds  []string `json:"pic_ids"`
			TextRaw string   `json:"text_raw"`
		} `json:"list"`
	} `json:"data"`
}
