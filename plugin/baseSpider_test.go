package plugin

import (
	"testing"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"github.com/hq-cml/spider-go/helper/log"
	"github.com/hq-cml/spider-go/basic"
	"io/ioutil"
	"fmt"
)

func TestParseATag(t *testing.T) {
	log.InitLog("", "debug")

	resp, err := http.DefaultClient.Get("https://www.360.cn/")       //360首页，UTF8编码，content-type: text/html，没有指明charset
	//resp, err := http.DefaultClient.Get("http://www.360.cn/n/10274.html")
	//resp, err := http.DefaultClient.Get("http://www.dygang.net/")    //电影港首页，gbk编码，content-type: text/html，没有指明charset
	//resp, err := http.DefaultClient.Get("https://www.jianshu.com") //简书首页，UTF8编码，content-type: text/html; charset=utf-8
	//resp, err := http.DefaultClient.Get("http://sd.360.cn/downloadoffline.html") //大文件

	//resp, err := http.DefaultClient.Get("http://www.360.cn/download") //这个里面有大量的二进制下载文件链接

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
	items, reqs, errors := parseForATag(httpResp, nil)

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
	for _, item := range items {
		t.Log((*item)["url"])
		t.Log((*item)["charset"])
		t.Log((*item)["body"])
	}

	t.Log("分析出的Error列表:", len(errors))
	for _, err := range errors {
		t.Log(err)
	}
}

func TestGoQuery(t *testing.T) {
	html := `<body>
				<div lang="zh">DIV1</div>
				<p>P1</p>
				<div lang="zh-cn">DIV2</div>
				<div lang="en">DIV3</div>
				<span>
					<div style="display:none;">DIV4</div>
					<div>DIV5</div>
				</span>
				<p>P2</p>
				<div></div>
			</body>
			`

	dom,err:=goquery.NewDocumentFromReader(strings.NewReader(html))
	if err!=nil{
		t.Fatal(err)
	}

	dom.Find("body").Each(func(i int, selection *goquery.Selection) {
		t.Log(i, selection.Text())
	})
}

func TestHttpRequest(t *testing.T) {
	firstHttpReq, err := http.NewRequest("GET", "https://www.360.cn/", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(firstHttpReq.URL.Scheme)
}