package basic
/*
 * 基本数据类型的定义
 */
import (
    "net/http"
)

/********************** 数据类型接口 *********************/
type DataIntfs interface {
    Valid() bool //数据是否有效
}

/********************** 请求体相关 **********************/
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

//实现DataIntfs接口
func (req *Request) Valid() bool {
    return req.httpReq != nil && req.httpReq.URL != nil
}

/**************** 响应体相关 *********************/
//响应体
type Response struct {
    httpResp *http.Request  //HTTP响应的指针
    depth    uint32         //深度
}

//惯例New函数，创建响应
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
func (resp *Response)Depth() unit32 {
    return resp.depth
}

//实现DataIntfs接口
func (resp *Response) Valid() bool {
    return resp.httpResp != nil && resp.httpResp.Body != nil
}

/********************** 条目相关 **********************/
type Item map[string]interface{}

//实现DataIntfs接口
func (item Item) Valid() bool {
    return item != nil
}














