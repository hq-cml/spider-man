package basic

import "errors"

//TODO 位置要移动到logic中
/*************************** 请求通道相关 ***************************/
func NewRequestChannel(capacity int) SpiderChannelIntfs {
	return &RequestChannel{
		capacity: capacity,
		reqCh:    make(chan Request, capacity),
	}
}
//实现SpiderChannelIntfs接口
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

/*************************** 响应通道相关 ***************************/
func NewResponseChannel(capacity int) SpiderChannelIntfs {
	return &ResponseChannel{
		capacity: capacity,
		respCh:   make(chan Response, capacity),
	}
}
//实现SpiderChannelIntfs接口
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

/*************************** 结果通道相关 ***************************/
func NewItemChannel(capacity int) SpiderChannelIntfs {
	return &ItemChannel{
		capacity: capacity,
		itemCh:  make(chan Item, capacity),
	}
}
//实现SpiderChannelIntfs接口
func (c *ItemChannel) Put(data interface{}) error {
	req, ok := data.(Item)
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
func NewErrorChannel(capacity int) SpiderChannelIntfs {
	return &ErrorChannel{
		capacity: capacity,
		errorCh:  make(chan SpiderError, capacity),
	}
}
//实现SpiderChannelIntfs接口
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
