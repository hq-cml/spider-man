package basic

import (
	"net/http"
	"fmt"
	"bytes"
)

/********************** Request 相关基本函数 **********************/
//New，创建Request
func NewRequest(httpReq *http.Request, depth int) *Request {
	return &Request{
		httpReq: httpReq,
		depth:   depth,
	}
}

//*Request实现Data接口
func (req *Request) Valid() bool {
	return req.httpReq != nil && req.httpReq.URL != nil
}

//获取请求值指针
func (req *Request) HttpReq() *http.Request {
	return req.httpReq
}

//获取深度值
func (req *Request) Depth() int {
	return req.depth
}

/************************** 响应体相关 **************************/
//New，创建响应
func NewResponse(body []byte, depth int, ct, url string) *Response {
	return &Response{
		Body        :body,
		Depth       :depth,
		ContentType :ct,
		ReqUrl      :url,
	}
}

//*Request实现Data接口
func (resp *Response) Valid() bool {
	return true
}

/*************************** 条目相关 ***************************/
func (e Item) Valid() bool {
	return e != nil
}

/*************************** 错误相关 ***************************/
//New
func NewSpiderErr(errType ErrorType, errMsg string) *SpiderError {
	return &SpiderError{
		errType: errType,
		errMsg:  errMsg,
	}
}

func (e *SpiderError) Type() ErrorType {
	return e.errType
}

//获得错误信息
func (e *SpiderError) Error() string {
	if e.fullErrMsg == "" {
		var buffer bytes.Buffer
		buffer.WriteString("Spider Error:")
		if e.errType != "" {
			buffer.WriteString(string(e.errType))
			buffer.WriteString(": ")
		}
		buffer.WriteString(e.errMsg)
		e.fullErrMsg = fmt.Sprintf("%s\n", buffer.String())
	}
	return e.fullErrMsg
}

