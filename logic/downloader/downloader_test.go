package downloader

import (
	"net/http"
	"testing"
	"github.com/hq-cml/spider-man/basic"
	"github.com/hq-cml/spider-man/helper/log"
)

func TestSkipUrl(t *testing.T) {
	log.InitLog("", "debug")

	dl := NewDownloader(nil)

	u, err := http.NewRequest(http.MethodGet, "https://www.360.cn/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req := basic.NewRequest(u, 0)
	t.Log(dl.skipBinFile(req))

	u, err = http.NewRequest(http.MethodGet, "https://dl.360safe.com/inst.exe", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = basic.NewRequest(u, 0)
	t.Log(dl.skipBinFile(req))


	u, err = http.NewRequest(http.MethodGet, "http://sd.360.cn/downloadoffline.html", nil) //伪静态大文件
	if err != nil {
		t.Fatal(err)
	}
	req = basic.NewRequest(u, 0)
	t.Log(dl.skipBinFile(req))
}

func TestSkipUrl2(t *testing.T) {
	log.InitLog("", "debug")
	dl := NewDownloader(nil)
	u, err := http.NewRequest(http.MethodGet, "http://bang.360.cn", nil)
	if err != nil {
		t.Fatal(err)
	}
	req := basic.NewRequest(u, 0)
	t.Log(dl.skipBinFile(req))
}

