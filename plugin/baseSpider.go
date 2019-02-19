package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hq-cml/spider-go/basic"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//*BaseSpider实现SpiderPlugin接口
type BaseSpider struct {
}

//New
func NewBaseSpider() basic.SpiderPlugin {
	return &BaseSpider{}
}

//生成HTTP客户端
func (b *BaseSpider) GenHttpClient() *http.Client {
	return &http.Client{}
}

//获得响应解析函数的序列
func (b *BaseSpider) GenResponseAnalysers() []basic.AnalyzeResponseFunc {
	analyzers := []basic.AnalyzeResponseFunc{
		parseForATag,
	}
	return analyzers
}

// 获得条目处理链的序列。
func (b *BaseSpider) GenItemProcessors() []basic.ProcessItemFunc {
	itemProcessors := []basic.ProcessItemFunc{
		processItem,
	}
	return itemProcessors
}

//响应解析函数。只解析“A”标签。
func parseForATag(httpResp *http.Response, grabDepth int) ([]*basic.Item, []*basic.Request, []error) {
	//仅支持返回码200的响应
	if httpResp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Unsupported status code %d. (httpResponse=%v)", httpResp))
		return nil, nil, []error{err}
	}

	//对响应做一些处理
	var reqUrl *url.URL = httpResp.Request.URL //记录下响应的请求（防止相对URL的问题）
	var httpRespBody io.ReadCloser = httpResp.Body
	//defer func() { //TODO 这一块应该放到框架中去？？？？
	//	if httpRespBody != nil {
	//		httpRespBody.Close()
	//	}
	//}()

	itemList := []*basic.Item{}
	requestList := []*basic.Request{}
	errs := make([]error, 0)
	//开始解析
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}

	uniqUrl := map[string]bool{}
	//查找“A”标签并提取链接地址
	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		// 前期过滤
		if !exists || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)
		// 暂不支持对Javascript代码的解析。
		if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
			aUrl, err := url.Parse(href)
			if err != nil {
				errs = append(errs, err)
				return
			}
			//保证是绝对URL(如果当前是相对URL，则将当前URL拼接到主URL上，保证是绝对URL)
			if !aUrl.IsAbs() {
				aUrl = reqUrl.ResolveReference(aUrl)
			}

			if _, ok := uniqUrl[aUrl.String()]; ok {  //去除重复的url
				return
			}
			uniqUrl[aUrl.String()] = true
			httpReq, err := http.NewRequest("GET", aUrl.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := basic.NewRequest(httpReq, grabDepth)
				//将新分析出来的请求，放入dataList
				requestList = append(requestList, req)
			}
		}
		text := strings.TrimSpace(sel.Text())
		//将新分析出来的Item，放入dataList
		if text != "" {
			imap := make(map[string]interface{})
			imap["parent_url"] = reqUrl
			imap["a.text"] = text
			imap["a.index"] = index
			item := basic.Item(imap)
			itemList = append(itemList, &item)
		}
	})
	return itemList, requestList, errs
}

// 条目处理器。
func processItem(item basic.Item) (result basic.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item!")
	}

	// 生成结果
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	//TODO 延时查看效果 ？？？
	time.Sleep(10 * time.Millisecond)
	s, _ := json.Marshal(result)
	fmt.Println("RRRRRRRRRRRRRRRRR", string(s))
	return result, nil
}
