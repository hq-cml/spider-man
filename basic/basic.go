package basic

import (
	"net/http"
	"errors"
)

/********************** Request 相关基本函数 **********************/
//New，创建Request
func NewRequest(httpReq *http.Request, depth uint32) *Request {
	return &Request{
		httpReq: httpReq,
		depth:   depth,
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
		depth:    depth,
	}
}

//获取响应体指针
func (resp *Response) HttpResp() *http.Response {
	return resp.httpResp
}

//获取响应的深度
func (resp *Response) Depth() uint32 {
	return resp.depth
}

//*Request实现DataIntfs接口
func (resp *Response) Valid() bool {
	return resp.httpResp != nil && resp.httpResp.Body != nil
}

/************************ 条目相关 ************************/
//实现EntryIntfs接口
func (e Entry) Valid() bool {
	return e != nil
}

/************************ 请求通道相关 ************************/
func NewRequestChannel(capacity int) SpiderChannelIntfs{
	return &RequestChannel{
		capacity: capacity,
		reqCh: make(chan Request, capacity),
	}
}
func (c *RequestChannel) Put(data interface{}) error {
	req, ok := data.(Request)
	if !ok {
		return errors.New("Wrong type")
	}

	c.reqCh <- req
	return nil
}
func (c *RequestChannel) Get() (interface{}, bool) {
	req, ok := <-c.reqCh
	return interface{}(req), ok
}
func (c *RequestChannel) Len() int {
	return len(c.reqCh)
}
func (c *RequestChannel) Cap() int {
	return c.capacity
}
func (c *RequestChannel) Close() {
	close(c.reqCh)
}

/************************ 响应通道相关 ************************/
func NewResponseChannel(capacity int) SpiderChannelIntfs{
	return &ResponseChannel{
		capacity: capacity,
		respCh: make(chan Response, capacity),
	}
}
func (r *ResponseChannel) Put(data interface{}) error {
	req, ok := data.(Response)
	if !ok {
		return errors.New("Wrong type")
	}

	r.respCh <- req
	return nil
}
func (r *ResponseChannel) Get() (interface{}, bool) {
	req, ok := <-r.respCh
	return interface{}(req), ok
}
func (r *ResponseChannel) Len() int {
	return len(r.respCh)
}
func (r *ResponseChannel) Cap() int {
	return r.capacity
}
func (c *ResponseChannel) Close() {
	close(c.respCh)
}

/************************ 结果通道相关 ************************/
func NewEntryChannel(capacity int) SpiderChannelIntfs{
	return &EntryChannel{
		capacity: capacity,
		entryCh: make(chan Entry, capacity),
	}
}
func (c *EntryChannel) Put(data interface{}) error {
	req, ok := data.(Entry)
	if !ok {
		return errors.New("Wrong type")
	}

	c.entryCh <- req
	return nil
}
func (c *EntryChannel) Get() (interface{}, bool) {
	req, ok := <-c.entryCh
	return interface{}(req), ok
}
func (c *EntryChannel) Len() int {
	return len(c.entryCh)
}
func (c *EntryChannel) Cap() int {
	return c.capacity
}
func (c *EntryChannel) Close() {
	close(c.entryCh)
}

/************************ 错误通道相关 ************************/
func NewErrorChannel(capacity int) SpiderChannelIntfs{
	return &ErrorChannel{
		capacity: capacity,
		errorCh: make(chan SpiderError, capacity),
	}
}
func (c *ErrorChannel) Put(data interface{}) error {
	req, ok := data.(SpiderError)
	if !ok {
		return errors.New("Wrong type")
	}

	c.errorCh <- req
	return nil
}
func (c *ErrorChannel) Get() (interface{}, bool) {
	req, ok := <-c.errorCh
	return interface{}(req), ok
}
func (c *ErrorChannel) Len() int {
	return len(c.errorCh)
}
func (c *ErrorChannel) Cap() int {
	return c.capacity
}
func (c *ErrorChannel) Close() {
	close(c.errorCh)
}