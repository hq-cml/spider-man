package plugin

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hq-cml/spider-man/basic"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//*BaseSpider实现SpiderPlugin接口
//一个最基础的插件，爬虫爬取季过后，直接进行关键字搜索出结果打印
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
	//客户端必须设置一个整体超时时间，否则随着时间推移，会把downloader全部卡死
	a := &http.Client{
		Transport: http.DefaultTransport,
		Timeout: time.Duration(basic.Conf.RequestTimeout) * time.Second,
	}

	//a.Transport.

	return a
}

//获得响应解析函数的序列
func (b *BaseSpider) GenResponseAnalysers() []basic.AnalyzeResponseFunc {
	analyzers := []basic.AnalyzeResponseFunc {
		//闭包
		func(httpResp *basic.Response) ([]*basic.Item, []*basic.Request, []error) {
			items, reqs, errs :=  parseForATag(httpResp, b.userData)
			return items, reqs, errs
		},
	}
	return analyzers
}

// 获得条目处理链的序列。
func (b *BaseSpider) GenItemProcessors() []basic.ProcessItemFunc {
	itemProcessors := []basic.ProcessItemFunc{
		processBaseItem,
	}
	return itemProcessors
}

/*
 * 分析函数
 * 分析出“A”标签,作为新的request
 * 分析出满足条件的结果作为item
 */
func parseForATag(httpResp *basic.Response, userData interface{}) ([]*basic.Item, []*basic.Request, []error) {

	//对响应做一些处理
	reqUrl, err := url.Parse(httpResp.ReqUrl) //记录下响应的请求（防止相对URL的问题）
	if err != nil {
		return nil, nil, []error{err}
	}
	var httpRespBody io.Reader

	itemList := []*basic.Item{}
	requestList := []*basic.Request{}
	errs := make([]error, 0)

	//网页编码智能判断, 非utf8 => utf8
	httpRespBody, contentType, err := convertCharset(httpResp)
	if err != nil {
		return nil, nil, []error{err}
	}

	//开始解析
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}

	//查找“A”标签并提取链接地址
	requestList, errs = findATagFromDoc(httpResp, reqUrl, doc)

	//关键字查找, 记录符合条件的body作为item
	//如果用户数据非空，则进行匹配校验，否则直接入item队列
	imap := make(map[string]interface{})
	imap["url"] = reqUrl.String()
	imap["charset"] = contentType
	imap["depth"] = httpResp.Depth
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


// 条目处理器。
func processBaseItem(item basic.Item) (result basic.Item, err error) {
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


