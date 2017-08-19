package channelmanager

import (
    "github.com/hq-cml/spider-go/basic"
    "sync"
)

/*
 * 中间件channel管理器
 * 框架中有四种类型的数据需要管道传递：请求、响应、条目、Error
 */

//通道管理器的状态的类型。
type ChannelManagerStatus uint8

const (
    CHANNEL_MANAGER_STATUS_UNINITIALIZED ChannelManagerStatus = iota //未初始化
    CHANNEL_MANAGER_STATUS_INITIALIZED                               //已完成初始化
    CHANNEL_MANAGER_STATUS_CLOSED                                    //已关闭
)

//状态码与状态名称映射字典
var statusNameMap = map[ChannelManagerStatus]string {
    CHANNEL_MANAGER_STATUS_UNINITIALIZED : "uninitialized",
    CHANNEL_MANAGER_STATUS_INITIALIZED : "inititalized",
    CHANNEL_MANAGER_STATUS_CLOSED : "closed",
}

//channel管理器接口
type ChannelManagerIntfs interface {
    // 初始化通道管理器。
    // 参数channelArgs代表通道参数的容器。
    // 参数reset指明是否重新初始化通道管理器。
    Init(channelParams basic.ChannelParams, reset bool) bool
    // 关闭通道管理器
    Close() bool
    // 获取请求传输通道。
    ReqChan() (chan basic.Request, error)
    // 获取响应传输通道。
    RespChan() (chan basic.Response, error)
    // 获取Entry传输通道。
    EntryChan() (chan basic.Entry, error)
    // 获取错误传输通道。
    ErrorChan() (chan error, error)
    // 获取通道管理器的状态。
    Status() ChannelManagerStatus
    // 获取摘要信息。
    Summary() string
}

//channel管理器实现类型
type ChannelManager struct {
    channelParams basic.ChannelParams  //通道长度
    reqCh         chan basic.Request   //请求通道
    respCh        chan basic.Response  //响应通道
    entryCh       chan basic.Entry     //entry通道
    errorCh       chan error           //错误通道
    status        ChannelManagerStatus //channel管理器状态
    rwmutex       sync.RWMutex         //读写锁
}
