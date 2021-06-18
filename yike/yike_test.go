package yike

import (
	"io/ioutil"
	"testing"
)

func TestPrecreate(t *testing.T) {
	bs, err := ioutil.ReadFile("C:/Users/Do/Downloads/金-01.jpg")
	if err != nil {
		t.Fatal(err)
	}
	yk := New(bs, "/tttt.jpg")
	resp, err := yk.precreate()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v\n", resp)
}

func TestUploadFile(t *testing.T) {
	bs, err := ioutil.ReadFile("C:/Users/Do/Downloads/33112314.png")
	if err != nil {
		t.Fatal(err)
	}

	yk := New(bs, "/tttt.jpg")
	err = yk.UploadFile()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("已上传文件\n")
}

func TestNew(t *testing.T) {
	bs, err := ioutil.ReadFile("C:/Users/Do/Downloads/33112314.png")
	if err != nil {
		t.Fatal(err)
	}
	yk := New(bs, "/test/tttt.jpg")
	t.Logf("分块的 MD5：%v\n", yk.BlockMD5List)

	err = yk.UploadFile()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("上传完成\n")
}
