package scheduler

/*
 * 调度器用到的的一些辅助的函数
 */
import (
    "github.com/hq-cml/spider-go/basic"
)

// 获取通道管理器持有的请求通道。
func (schdl *Scheduler) getReqestChan() chan basic.Request {
    requestChan, err := schdl.channelManager.ReqChan()
    if err != nil {
        panic(err)
    }
    return requestChan
}

// 获取通道管理器持有的响应通道。
func (schdl *Scheduler) getResponseChan() chan basic.Response {
    respChan, err := schdl.channelManager.RespChan()
    if err != nil {
        panic(err)
    }
    return respChan
}

// 获取通道管理器持有的条目通道。
func (schdl *Scheduler) getItemChan() chan basic.Item {
    itemChan, err := schdl.channelManager.ItemChan()
    if err != nil {
        panic(err)
    }
    return itemChan
}

// 获取通道管理器持有的错误通道。
func (schdl *Scheduler) getErrorChan() chan error {
    errorChan, err := schdl.channelManager.ErrorChan()
    if err != nil {
        panic(err)
    }
    return errorChan
}