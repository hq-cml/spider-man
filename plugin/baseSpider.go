package plugin

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hq-cml/spider-go/basic"
	"io"
	"net/http"
	"net/url"
	"strings"
	"io/ioutil"
	"bytes"
	"github.com/hq-cml/spider-go/helper/log"
	"golang.org/x/net/html/charset"
	"bufio"
	"golang.org/x/text/transform"
)

//*BaseSpider实现SpiderPlugin接口
type BaseSpider struct {
	userData interface{}
}

//New
func NewBaseSpider(v interface{}) basic.SpiderPlugin {
	return &BaseSpider{
		userData: v,
	}
}

//生成HTTP客户端
func (b *BaseSpider) GenHttpClient() *http.Client {
	return &http.Client{}
}

//获得响应解析函数的序列
func (b *BaseSpider) GenResponseAnalysers() []basic.AnalyzeResponseFunc {
	analyzers := []basic.AnalyzeResponseFunc {
		//闭包
		func(httpResp *http.Response, respDepth int) ([]*basic.Item, []*basic.Request, []error) {
			items, reqs, errs :=  parseForATag(httpResp, respDepth, b.userData)
			return items, reqs, errs
		},
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

/*
 * 分析函数
 * 分析出“A”标签,作为新的request
 * 分析出满足条件的结果作为item
 */
func parseForATag(httpResp *http.Response, grabDepth int, userData interface{}) ([]*basic.Item, []*basic.Request, []error) {
	//这个地方其实已经没必要这么处理，因为在downloader.go中已经关闭了真正的Body
	//只是为了保证使用方式的统一，这也真是NopCloser需要的效果
	p, _ := ioutil.ReadAll(httpResp.Body)
	httpResp.Body.Close()
	httpResp.Body = ioutil.NopCloser(bytes.NewBuffer(p))

	//对响应做一些处理
	var reqUrl *url.URL = httpResp.Request.URL //记录下响应的请求（防止相对URL的问题）
	var httpRespBody io.Reader
	var err error

	itemList := []*basic.Item{}
	requestList := []*basic.Request{}
	errs := make([]error, 0)

	//网页编码智能判断, 非utf8 => utf8
	httpRespBody, contentType, err := convertCharset(httpResp, p)
	if err != nil {
		return nil, nil, []error{err}
	}

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
			httpReq, err := http.NewRequest(http.MethodGet, aUrl.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				//将新分析出来的请求，深度+1，放入dataList
				req := basic.NewRequest(httpReq, grabDepth + 1)
				requestList = append(requestList, req)
			}
		}
		//text := strings.TrimSpace(sel.Text())
		////将新分析出来的Item，放入dataList
		//if text != "" {
		//	imap := make(map[string]interface{})
		//	imap["href"] = href
		//	imap["text"] = text
		//	imap["index"] = index
		//	item := basic.Item(imap)
		//	itemList = append(itemList, &item)
		//}
	})

	//关键字查找, 记录符合条件的body作为item
	//如果用户数据非空，则进行匹配校验，否则直接入item队列
	imap := make(map[string]interface{})
	imap["url"] = reqUrl.String()
	imap["charset"] = contentType
	imap["depth"] = grabDepth
	body := doc.Find("body").Text()
	imap["body"] = body
	item := basic.Item(imap)
	if userData != nil {
		searchContent, ok := userData.(string)
		if !ok {
			err := errors.New(fmt.Sprintf("Unsupported userData=%v", userData))
			return nil, nil, []error{err}
		}

		if strings.Contains(body, searchContent) {
			itemList = append(itemList, &item)
		}
	} else {
		itemList = append(itemList, &item)
	}

	return itemList, requestList, errs
}

//识别编码并转换到utf8
func convertCharset(httpResp *http.Response, body []byte) (httpRespBody io.Reader, orgCharset string, err error) {
	//先尝试从http header的content-type中直接猜测
	contentType := httpResp.Header.Get("content-type")
	switch {
	case strings.Contains(strings.ToLower(contentType), "utf8"),
		strings.Contains(strings.ToLower(contentType), "utf-8"):
		httpRespBody = bytes.NewBuffer(body)
		return httpRespBody, "utf8", nil
	}

	//利用golang.org/x/net/html/charset包提供的方法开始猜
	buf, err := bufio.NewReader(bytes.NewBuffer(body)).Peek(1024)
	if err != nil {
		panic(err)
	}
	encoding, charset, _ := charset.DetermineEncoding(buf, contentType)

	//利用http.DetectContentType进行猜测
	switch {
	case strings.Contains(strings.ToLower(charset), "utf8"),
		strings.Contains(strings.ToLower(charset), "utf-8"):
		httpRespBody = bytes.NewBuffer(body)
		log.Debugln("Determine charset: utf8. Url:", httpResp.Request.URL.String(), ". Content-Type:", contentType)
		return httpRespBody, "utf8", nil

	default:
		//需要转码
		log.Debugln("Guess charset:", charset,". Url:", httpResp.Request.URL.String(), ". Content-Type:", contentType)
		utf8Body := transform.NewReader(bytes.NewBuffer(body), encoding.NewDecoder())
		return utf8Body, "", nil
	}
}

// 条目处理器。
func processItem(item basic.Item) (result basic.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item!")
	}

	//生成结果
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}

	fmt.Println("深度: ", result["depth"], "结果：", result["url"])
	return
}


