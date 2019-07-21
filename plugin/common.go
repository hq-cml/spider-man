package plugin

import (
	"github.com/hq-cml/spider-man/basic"
	"io"
	"strings"
	"bytes"
	"bufio"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
	"github.com/hq-cml/spider-man/helper/log"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"net/http"
)

//识别编码并转换到utf8
func convertCharset(httpResp *basic.Response) (httpRespBody io.Reader, orgCharset string, err error) {
	//先尝试从http header的content-type中直接猜测
	//contentType := httpResp.Header.Get("content-type")
	switch {
	case strings.Contains(strings.ToLower(httpResp.ContentType), "utf8"),
		strings.Contains(strings.ToLower(httpResp.ContentType), "utf-8"):
		httpRespBody = bytes.NewBuffer(httpResp.Body)
		return httpRespBody, "utf8", nil
	}

	//利用golang.org/x/net/html/charset包提供的方法开始猜
	buf, err := bufio.NewReader(bytes.NewBuffer(httpResp.Body)).Peek(1024)
	if err != nil {
		panic(err)
	}
	encoding, charSet, _ := charset.DetermineEncoding(buf, httpResp.ContentType)

	//利用http.DetectContentType进行猜测
	switch {
	case strings.Contains(strings.ToLower(charSet), "utf8"),
		strings.Contains(strings.ToLower(charSet), "utf-8"):
		httpRespBody = bytes.NewBuffer(httpResp.Body)
		log.Debugln("Determine charset: utf8. Url:", httpResp.ReqUrl, ". Content-Type:", httpResp.ContentType)
		return httpRespBody, "utf8", nil

	default:
		//需要转码
		log.Debugln("Guess charset:", charSet,". Url:", httpResp.ReqUrl, ". Content-Type:", httpResp.ContentType)
		utf8Body := transform.NewReader(bytes.NewBuffer(httpResp.Body), encoding.NewDecoder())
		return utf8Body, charSet, nil
	}
}

//从document当中扫出全部的链接Tag，然后拼成新请求
func findATagFromDoc(httpResp *basic.Response, reqUrl *url.URL, doc *goquery.Document,
	requestList *[]*basic.Request, errs []error) {

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
		if href == "" || strings.HasPrefix(lowerHref, "javascript") {
			return
		}

		aUrl, err := url.Parse(href)
		if err != nil {
			errs = append(errs, err)
			return
		}
		//保证是绝对URL(如果当前是相对URL，则将当前URL拼接到主URL上，保证是绝对URL)
		if !aUrl.IsAbs() {
			aUrl = reqUrl.ResolveReference(aUrl)
		}

		//去除本页面内部#干扰和重复的url
		uurl := aUrl.String()
		uurl = strings.Split(uurl, "#")[0]
		uurl = strings.TrimRight(uurl, "/")
		if _, ok := uniqUrl[uurl]; ok {
			return
		}
		uniqUrl[uurl] = true
		httpReq, err := http.NewRequest(http.MethodGet, uurl, nil)
		if err != nil {
			errs = append(errs, err)
		} else {
			//将新分析出来的请求，深度+1，放入dataList
			req := basic.NewRequest(httpReq, httpResp.Depth + 1)
			*requestList = append(*requestList, req)
		}
	})
}