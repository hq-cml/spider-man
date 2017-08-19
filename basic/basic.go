package basic

import (
    "net/http"
)

/********************** Request 相关基本函数 **********************/
//New，创建Request
func NewRequest(httpReq *http.Request, depth uint32) *Request {
    return &Request{
        httpReq: httpReq,
        depth: depth,
    }
}

//获取请求值指针
func (req *Request) HttpReq() *http.Request {
    return req.httpReq
}

//获取深度值
func (req *Request) Depth() uint32 {
    return req.depth
}

//*Request实现DataIntfs接口
func (req *Request) Valid() bool {
    return req.httpReq != nil && req.httpReq.URL != nil
}

/*********************** 响应体相关 ***********************/
//New，创建响应
func NewResponse(httpResp *http.Response, depth uint32) *Response {
    return &Response{
        httpResp: httpResp,
        depth: depth,
    }
}

//获取响应体指针
func (resp *Response)HttpResp() *http.Response {
    return resp.httpResp
}

//获取响应的深度
func (resp *Response)Depth() uint32 {
    return resp.depth
}

//*Request实现DataIntfs接口
func (resp *Response) Valid() bool {
    return resp.httpResp != nil && resp.httpResp.Body != nil
}

/************************ 条目相关 ************************/
//实现DataIntfs接口
func (e Entry) Valid() bool {
    return e != nil
}