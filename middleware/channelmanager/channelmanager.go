package channelmanager

import "github.com/hq-cml/spider-go/basic"

/*
 * 中间件channel管理器
 *
 */

//通道管理器的状态的类型。
type ChannelManagerStatus uint8

const (
    CHANNEL_MANAGER_STATUS_UNINITIALIZED ChannelManagerStatus = iota //未初始化
    CHANNEL_MANAGER_STATUS_INITIALIZED                               //已完成初始化
    CHANNEL_MANAGER_STATUS_CLOSED                                    //已关闭
)

//channel管理器接口
type ChannelManagerIntfs interface {
    // 初始化通道管理器。
    // 参数channelArgs代表通道参数的容器。
    // 参数reset指明是否重新初始化通道管理器。
    Init(channelLen uint, reset bool) bool
    // 关闭通道管理器
    Close() bool
    // 获取请求传输通道。
    ReqChan() (chan basic.Request, error)
    // 获取响应传输通道。
    RespChan() (chan basic.Response, error)
    // 获取Item传输通道。
    ItemChan() (chan basic.Item, error)
    // 获取错误传输通道。
    ErrorChan() (chan error, error)
    // 获取通道管理器的状态。
    Status() ChannelManagerStatus
    // 获取摘要信息。
    Summary() string
}