package plugin

import (
	"testing"
	"net/http"
	"strings"
	"github.com/hq-cml/spider-man/helper/log"
	"github.com/hq-cml/spider-man/basic"
	"io/ioutil"
	"fmt"
	"github.com/hq-cml/spider-man/helper/util"
)

func TestParse360NewsPage(t *testing.T) {
	log.InitLog("", "debug")

	//resp, err := http.DefaultClient.Get("http://www.360.cn/n/10758.html")
	//resp, err := http.DefaultClient.Get("http://www.360.cn/n/10759.html")
	resp, err := http.DefaultClient.Get("http://www.360.cn/news.html")

	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("Content-Type----", resp.Header.Get("content-type"))
	fmt.Println("Status-Code----", resp.StatusCode)
	fmt.Println("Content-Length--", len(body))

	httpResp := basic.NewResponse(
		body,
		0,
		resp.Header.Get("content-type"),
		resp.Request.URL.String(),
	)
	items, reqs, errors := parse360NewsPage(httpResp)

	t.Logf("分析出的URL列表(%d):\n", len(reqs))
	m := map[string]bool{}

	for _, req := range reqs {
		url := req.HttpReq().URL.String()
		url = strings.Split(url, "#")[0]
		url = strings.TrimRight(url, "/")
		if _, ok := m[url]; ok {
			continue
		}
		m[url] = true
	}
	t.Logf("去重后的URL列表(%d):\n", len(m))
	for url,_ := range m {
		t.Logf("URL: %s", url)
	}

	t.Log("分析出的Item列表:", len(items))
	t.Log(util.JsonEncode(items))

	t.Log("分析出的Error列表:", len(errors))
	for _, err := range errors {
		t.Log(err)
	}

	//发送数据到spider-engine
	//if len(items) > 0 {
	//	item := items[0]
	//	t.Log("数据发送：", postOneNews(*item, "http://192.168.110.133:9528/sp_db/360news/10759"))
	//}

}

