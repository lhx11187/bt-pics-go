package yike

// PreResp 响应
type PreResp struct {
	ReturnType int    `json:"return_type"`
	Uploadid   string `json:"uploadid"`
	Errno      int    `json:"errno"`
}

// UpResp 上传分段的响应
type UpResp struct {
	Md5      string `json:"md5"`
	Partseq  string `json:"partseq"`
	Uploadid string `json:"uploadid"`
}

// FilesResp 文件列表
type FilesResp struct {
	Cursor string `json:"cursor"`
	Errno  int    `json:"errno"`
	List   []struct {
		Fsid int64 `json:"fsid"`
	} `json:"list"`
}
