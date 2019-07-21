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