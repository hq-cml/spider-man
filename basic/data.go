package basic

import (
    "net/http"
)

//请求体（为了避免零值填充和实例复制，结构体成员尽量用指针）
type Request struct {
    httpReq *http.Request   //HTTP请求的指针
    depth   uint32          //请求深度，初始请求深度是0，然后逐渐递增
}

//惯例New函数，创建请求
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