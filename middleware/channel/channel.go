package channel

import (
	"errors"
	"github.com/hq-cml/spider-go/basic"
)

//请求通道
type RequestChannel struct {
	capacity int
	reqCh    chan basic.Request
}

//响应通道
type ResponseChannel struct {
	capacity int
	respCh   chan basic.Response
}

//结果通道
type ItemChannel struct {
	capacity int
	itemCh   chan basic.Item
}

//错误通道
type ErrorChannel struct {
	capacity int
	errorCh  chan basic.SpiderError
}

/*************************** 请求通道相关 ***************************/
func NewRequestChannel(capacity int) basic.SpiderChannel {
	return &RequestChannel{
		capacity: capacity,
		reqCh:    make(chan basic.Request, capacity),
	}
}
//实现SpiderChannelIntfs接口
func (c *RequestChannel) Put(data interface{}) error {
	req, ok := data.(basic.Request)
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

/*************************** 响应通道相关 ***************************/
func NewResponseChannel(capacity int) basic.SpiderChannel {
	return &ResponseChannel{
		capacity: capacity,
		respCh:   make(chan basic.Response, capacity),
	}
}
//实现SpiderChannelIntfs接口
func (r *ResponseChannel) Put(data interface{}) error {
	req, ok := data.(basic.Response)
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

/*************************** 结果通道相关 ***************************/
func NewItemChannel(capacity int) basic.SpiderChannel {
	return &ItemChannel{
		capacity: capacity,
		itemCh:  make(chan basic.Item, capacity),
	}
}
//实现SpiderChannelIntfs接口
func (c *ItemChannel) Put(data interface{}) error {
	req, ok := data.(basic.Item)
	if !ok {
		return errors.New("Wrong type")
	}

	c.itemCh <- req
	return nil
}
func (c *ItemChannel) Get() (interface{}, bool) {
	req, ok := <-c.itemCh
	return interface{}(req), ok
}
func (c *ItemChannel) Len() int {
	return len(c.itemCh)
}
func (c *ItemChannel) Cap() int {
	return c.capacity
}
func (c *ItemChannel) Close() {
	close(c.itemCh)
}

/*************************** 错误通道相关 ***************************/
func NewErrorChannel(capacity int) basic.SpiderChannel {
	return &ErrorChannel{
		capacity: capacity,
		errorCh:  make(chan basic.SpiderError, capacity),
	}
}
//实现SpiderChannelIntfs接口
func (c *ErrorChannel) Put(data interface{}) error {
	req, ok := data.(basic.SpiderError)
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
