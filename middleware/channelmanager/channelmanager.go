package channelmanager

import (
    "github.com/hq-cml/spider-go/basic"
    "errors"
    "fmt"
)

//惯例New函数
func NewChannelManager(channelLen uint) ChannelManagerIntfs {
    chanman := &ChannelManager{}
    chanman.Init(channelLen, true)
    return chanman
}

// 检查状态。在获取通道的时候，通道管理器应处于已初始化状态。
// 如果通道管理器未处于已初始化状态，那么本方法将会返回一个非nil的错误值。
func (chanman *ChannelManager) checkStatus() error {
    if chanman.status == CHANNEL_MANAGER_STATUS_INITIALIZED {
        return nil
    }
    statusName, ok := statusNameMap[chanman.status]
    if !ok {
        statusName = fmt.Sprintf("%d", chanman.status)
    }
    errMsg := fmt.Sprintf("the undesirable status of channel manager :%s!\n", statusName)
    return errors.New(errMsg)
}

//*ChannelManager实现ChannelManagerIntfs接口
//Init方法
func (chanman *ChannelManager) Init(channelLen uint, reset bool) bool {
    if channelLen == 0 {
        panic(errors.New("The channel length is invalid!"))
    }
    //写锁保护
    chanman.rwmutex.Lock()
    defer chanman.rwmutex.Unlock()

    //避免重复初始化
    if chanman.status == CHANNEL_MANAGER_STATUS_INITIALIZED && reset != true {
        return false
    }
    chanman.channelLen = channelLen
    chanman.reqCh = make(chan basic.Request, channelLen)
    chanman.respCh = make(chan basic.Response, channelLen)
    chanman.itemCh = make(chan basic.Item, channelLen)
    chanman.errorCh = make(chan error, channelLen)
    chanman.status = CHANNEL_MANAGER_STATUS_INITIALIZED

    return true
}

//close关闭
func (chanman *ChannelManager) Close() bool {
    //写锁保护
    chanman.rwmutex.Lock()
    defer chanman.rwmutex.Unlock()

    if chanman.status != CHANNEL_MANAGER_STATUS_INITIALIZED {
        return false
    }

    close(chanman.reqCh)
    close(chanman.respCh)
    close(chanman.itemCh)
    close(chanman.errorCh)
    chanman.status = CHANNEL_MANAGER_STATUS_CLOSED

    return true
}

//获取request通道
func (chanman *ChannelManager) ReqChan() (chan basic.Request, error) {
    //读锁保护
    chanman.rwmutex.RLock()
    defer chanman.rwmutex.RUnlock()
    if err := chanman.checkStatus(); err != nil {
        return nil, err
    }
    return chanman.reqCh, nil
}

//获取response通道
func (chanman *ChannelManager) RespChan() (chan basic.Response, error) {
    //读锁保护
    chanman.rwmutex.RLock()
    defer chanman.rwmutex.RUnlock()
    if err := chanman.checkStatus(); err != nil {
        return nil, err
    }
    return chanman.respCh, nil
}

//获取item通道
func (chanman *ChannelManager) ItemChan() (chan basic.Item, error) {
    //读锁保护
    chanman.rwmutex.RLock()
    defer chanman.rwmutex.RUnlock()
    if err := chanman.checkStatus(); err != nil {
        return nil, err
    }
    return chanman.itemCh, nil
}

//获取error通道
func (chanman *ChannelManager) ErrorChan() (chan error, error) {
    //读锁保护
    chanman.rwmutex.RLock()
    defer chanman.rwmutex.RUnlock()
    if err := chanman.checkStatus(); err != nil {
        return nil, err
    }
    return chanman.errorCh, nil
}

//摘要方法
func (chanman *ChannelManager) Summary() string {
    //模板
    chanmanSummaryTemplate := "status: %s, " +
    "requestChannel: %d/%d, " +
    "responseChannel: %d/%d, " +
    "itemChannel: %d/%d, " +
    "errorChannel: %d/%d"

    summary := fmt.Sprintf(chanmanSummaryTemplate,
        statusNameMap[chanman.status],
        len(chanman.reqCh), cap(chanman.reqCh),
        len(chanman.respCh), cap(chanman.respCh),
        len(chanman.itemCh), cap(chanman.itemCh),
        len(chanman.errorCh), cap(chanman.errorCh))
    return summary
}












