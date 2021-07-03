package yike

import (
	"bt-pics-go/client"
	"bt-pics-go/comm"
	"bt-pics-go/conf"
	"bt-pics-go/logger"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const (
	// 按 4MB 分割文件上传
	splitSize = 4 * 1024 * 1024
	// 取前 256 KB  字节计算 MD5
	md5Size = 256 * 1024
)

// YkFile 一刻相册的文件类型
type YkFile struct {
	Origin       []byte   // 文件的二进制数据
	BlockList    [][]byte // 按每 4MB 分块文件，得到的二进制数据
	BlockMD5List []string // 每个分块的 MD5
	BlockMD5Str  string   // 每个分块的 MD5 数组被转为字符串
	Path         string   // 文件被保存到的远程目录，如"/filename.jpg"
	Isdir        int
	Size         int
	SliceMd5     string
	ContentMd5   string

	// 可选
	LocalCtime int64 // 创建时间
}

var (
	// 预处理数据的 URL
	precreateURL = "https://photo.baidu.com/youai/file/v1/precreate?clienttype=70&bdstoken=%s"
	// 上传分段的 URL
	superfileURL = "https://c3.pcs.baidu.com/rest/2.0/pcs/superfile2?method=upload&app_id=16051585" +
		"&channel=chunlei&clienttype=70&web=1&logid=MTYyNDAwODkyNzY1NTAuNzEyMjQyOTExODk0OTE1" +
		"&path=%s&uploadid=%s&partseq=%d"
	createURL = "https://photo.baidu.com/youai/file/v1/create?clienttype=70&bdstoken=%s"
	// 列出文件
	listURL = "https://photo.baidu.com/youai/file/v1/list?clienttype=70&" +
		"bdstoken=%s&need_thumbnail=1&need_filter_hidden=0"
	// 删除文件
	delURL = "https://photo.baidu.com/youai/file/v1/delete?clienttype=70&bdstoken=%s&fsid_list=%s"
	// 请求头
	headers map[string]string
)

func init() {
	precreateURL = fmt.Sprintf(precreateURL, conf.Conf.Yike.Bdstoken)
	createURL = fmt.Sprintf(createURL, conf.Conf.Yike.Bdstoken)
	listURL = fmt.Sprintf(listURL, conf.Conf.Yike.Bdstoken)

	headers = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/91.0.4472.106 Safari/537.36",
		"Origin":  "https://photo.baidu.com",
		"Referer": "https://photo.baidu.com/photo/web/home",
		"Cookie":  conf.Conf.Yike.Cookie,
	}
}

// SendmAlbum 发送图集
func SendmAlbum(album comm.Album) error {
	for _, u := range album.URLs {
		// 先下载文件
		bs, err := client.Client.Get(u, album.Header)
		if err != nil {
			return fmt.Errorf("下载文件'%s'出错：%s\n", u, err)
		}

		// 发送到一刻相册
		filename := filepath.Base(u)
		if i := strings.Index(filename, "?"); i >= 0 {
			filename = filename[:i]
		}
		yk := New(bs, fmt.Sprintf("/%s/%s", album.ID, filename), album.Created)
		yk.Origin = nil
		yk.BlockList = nil
		err = yk.UploadFile()
		if err != nil {
			return fmt.Errorf("发送文件(%d KB)到一刻相册时出错：%s\n", len(bs)/1024, err)
		}
	}
	return nil
}

// DelAll 删除所有文件
func DelAll() error {
	for {
		// 列出文件
		bs, err := client.Client.Get(listURL, headers)
		if err != nil {
			return fmt.Errorf("列出文件出错：%w", err)
		}
		var files FilesResp
		err = json.Unmarshal(bs, &files)
		if err != nil {
			return fmt.Errorf("解析文件列表出错：%w", err)
		}
		if files.Errno != 0 {
			return fmt.Errorf("列出文件失败：%s\n", string(bs))
		}

		// 删除文件
		fidList := make([]int64, len(files.List))
		for i, f := range files.List {
			fidList[i] = f.Fsid
		}
		bs, err = json.Marshal(fidList)
		if err != nil {
			return fmt.Errorf("序列化文件的 ID 列表时出错：%w", err)
		}
		u := fmt.Sprintf(delURL, conf.Conf.Yike.Bdstoken, string(bs))
		bs, err = client.Client.Get(u, headers)
		if err != nil {
			return fmt.Errorf("删除文件出错：%w", err)
		}
		var resp PreResp
		err = json.Unmarshal(bs, &resp)
		if err != nil {
			return fmt.Errorf("解析删除文件的响应时出错：%w", err)
		}
		if resp.Errno == 2 {
			break
		}
		if resp.Errno != 0 {
			logger.Error.Printf("删除文件失败：%s\n", string(bs))
			// return fmt.Errorf("删除文件失败：%s\n", string(bs))
		}

		logger.Info.Printf("已删除此页，将继续删除下页\n")
		time.Sleep(1 * time.Second)
	}
	logger.Info.Printf("已尝试删除所有文件，若还有遗漏，可再次运行本程序\n")
	return nil
}

// New 创建一刻文件的实例
// createdTime 为 0 时，将自动设为 Unix 时间戳（秒）
func New(data []byte, remotePath string, createdTime int64) *YkFile {
	// 文件将被分成的段数
	blockNum := int(math.Ceil(float64(len(data)) / float64(splitSize)))

	// 当文件大小小于 md5Size 时，两个 MD5 相同
	var contentMd5 = md5.Sum(data)
	var sliceMd5 = contentMd5
	if len(data) > md5Size {
		sliceMd5 = md5.Sum(data[:md5Size])
	}

	// 文件对象，将返回
	ykFile := YkFile{
		Origin:       data,
		BlockList:    make([][]byte, blockNum),
		BlockMD5List: make([]string, blockNum),
		Isdir:        0,
		Path:         remotePath,
		Size:         len(data),
		ContentMd5:   fmt.Sprintf("%x", contentMd5),
		SliceMd5:     fmt.Sprintf("%x", sliceMd5),
	}
	// 其它属性
	sec := createdTime
	if sec == 0 {
		sec = time.Now().Unix()
	}
	ykFile.LocalCtime = sec
	// 将文件分段
	i := 0
	for pos := 0; i < blockNum; pos += splitSize {
		var tmp []byte
		// 最后一个分段为 [pos:]，其它分段为 [pos : pos+splitSize]
		if i < blockNum-1 {
			tmp = data[pos : pos+splitSize]
		} else {
			tmp = data[pos:]
		}
		// 添加分段
		ykFile.BlockList[i] = tmp
		// 保存 MD5
		ykFile.BlockMD5List[i] = fmt.Sprintf("%x", md5.Sum(tmp))
		i++
	}

	// 将 md5 的数组转为字符串
	md5BS, _ := json.Marshal(ykFile.BlockMD5List)
	ykFile.BlockMD5Str = string(md5BS)

	return &ykFile
}

// UploadFile 上传文件到一刻相册
func (yk *YkFile) UploadFile() error {
	resp, err := yk.precreate()
	if err != nil {
		return err
	}
	// type 为 1，表示云端没有该文件，需要上传
	if resp.ReturnType == 1 {
		// 上传
		err = yk.superfile(resp)
		if err != nil {
			return err
		}
		err = yk.create(resp.Uploadid)
		return err
	} else if resp.ReturnType == 2 || resp.ReturnType == 3 {
		// type 为 2或3，都表示云端已有改文件，可以“秒传”
		return nil
	}
	return fmt.Errorf("未知的 type：%+v", resp)
}

// 预处理数据文件
func (yk *YkFile) precreate() (*PreResp, error) {
	// 创建表单
	// "rtype"的值需要为"3"（覆盖文件）
	form := url.Values{}
	form.Add("autoinit", "1")
	form.Add("isdir", fmt.Sprintf("%d", yk.Isdir))
	form.Add("rtype", "3")
	form.Add("ctype", "11")
	form.Add("path", yk.Path)

	form.Add("content-md5", yk.ContentMd5)
	form.Add("size", fmt.Sprintf("%d", yk.Size))
	form.Add("slice-md5", yk.SliceMd5)
	form.Add("block_list", yk.BlockMD5Str)
	form.Add("local_ctime", fmt.Sprintf("%d", yk.LocalCtime))
	// form.Add("local_mtime", fmt.Sprintf("%d", time.Now().Unix()))

	// 发送表单
	bs, err := client.Client.PostForm(precreateURL, form.Encode(), headers)
	if err != nil {
		return nil, err
	}
	// 解析
	var resp PreResp
	err = json.Unmarshal(bs, &resp)
	if err != nil {
		return &resp, fmt.Errorf("解析 precreate 的响应出错：%w ==> %s", err, string(bs))
	}
	// 响应不符合要求
	if resp.Errno != 0 {
		return &resp, fmt.Errorf("预上传分段失败：%s", string(bs))
	}
	return &resp, nil
}

// 分段上传
// @see https://www.coder.work/article/207920
func (yk *YkFile) superfile(resp *PreResp) error {
	for i := 0; i < len(yk.BlockList); i++ {
		// process buf
		// 上传片段
		u := fmt.Sprintf(superfileURL, yk.Path, resp.Uploadid, i)
		file := map[string]interface{}{"file": yk.BlockList[i]}
		bs, err := client.Client.PostFiles(u, file, nil, headers)
		if err != nil {
			return fmt.Errorf("上传分段出错：%w ==> %s", err, string(bs))
		}

		// 解析结果
		var upResp UpResp
		err = json.Unmarshal(bs, &upResp)
		if err != nil {
			return fmt.Errorf("解析上传分段的响应出错：%w ==> %s", err, string(bs))
		}

		// 响应不符合要求
		if upResp.Uploadid != resp.Uploadid {
			return fmt.Errorf("上传分段失败：%s", string(bs))
		}
	}
	return nil
}

// 根据上传的分段，生成文件
func (yk *YkFile) create(uploadid string) error {
	// 创建表单
	form := url.Values{}
	form.Add("isdir", fmt.Sprintf("%d", yk.Isdir))
	form.Add("rtype", "3")
	form.Add("ctype", "11")
	form.Add("path", yk.Path)

	form.Add("content-md5", yk.ContentMd5)
	form.Add("size", fmt.Sprintf("%d", yk.Size))
	form.Add("uploadid", uploadid)
	form.Add("block_list", yk.BlockMD5Str)

	_, err := client.Client.PostForm(createURL, form.Encode(), headers)
	return err
}
