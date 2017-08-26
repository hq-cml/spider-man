package channelmanager

import (
	"errors"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
)

//New
func NewChannelManager(chp basic.ChannelParams) ChannelManagerIntfs {
	chm := &ChannelManager{}
	chm.Init(chp, true)
	return chm
}

// 检查状态，保证在获取通道的时候，通道管理器应处于已初始化状态。
func (chm *ChannelManager) checkStatus() error {
	if chm.status == CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	statusName, ok := statusNameMap[chm.status]
	if !ok {
		statusName = fmt.Sprintf("%d", chm.status)
	}
	errMsg := fmt.Sprintf("the undesirable status of channel manager :%s!\n", statusName)
	return errors.New(errMsg)
}

//*ChannelManager实现ChannelManagerIntfs接口
//Init方法
func (chm *ChannelManager) Init(chp basic.ChannelParams, reset bool) bool {
	if chp.Check() != nil {
		panic(errors.New("The channel length is invalid!"))
	}
	//写锁保护
	chm.rwmutex.Lock()
	defer chm.rwmutex.Unlock()

	//避免重复初始化
	if chm.status == CHANNEL_MANAGER_STATUS_INITIALIZED && reset != true {
		return false
	}
	chm.channelParams = chp
	chm.reqCh = make(chan basic.Request, chp.ReqChanLen())
	chm.respCh = make(chan basic.Response, chp.RespChanLen())
	chm.entryCh = make(chan basic.Entry, chp.EntryChanLen())
	chm.errorCh = make(chan error, chp.ErrorChanLen())
	chm.status = CHANNEL_MANAGER_STATUS_INITIALIZED

	return true
}

//close关闭
func (chm *ChannelManager) Close() bool {
	//写锁保护
	chm.rwmutex.Lock()
	defer chm.rwmutex.Unlock()

	if chm.status != CHANNEL_MANAGER_STATUS_INITIALIZED {
		return false
	}

	close(chm.reqCh)
	close(chm.respCh)
	close(chm.entryCh)
	close(chm.errorCh)
	chm.status = CHANNEL_MANAGER_STATUS_CLOSED

	return true
}

//获取request通道
func (chm *ChannelManager) ReqChan() (chan basic.Request, error) {
	//读锁保护
	chm.rwmutex.RLock()
	defer chm.rwmutex.RUnlock()
	if err := chm.checkStatus(); err != nil {
		return nil, err
	}
	return chm.reqCh, nil
}

//获取response通道
func (chm *ChannelManager) RespChan() (chan basic.Response, error) {
	//读锁保护
	chm.rwmutex.RLock()
	defer chm.rwmutex.RUnlock()
	if err := chm.checkStatus(); err != nil {
		return nil, err
	}
	return chm.respCh, nil
}

//获取entry通道
func (chm *ChannelManager) EntryChan() (chan basic.Entry, error) {
	//读锁保护
	chm.rwmutex.RLock()
	defer chm.rwmutex.RUnlock()
	if err := chm.checkStatus(); err != nil {
		return nil, err
	}
	return chm.entryCh, nil
}

//获取error通道
func (chm *ChannelManager) ErrorChan() (chan error, error) {
	//读锁保护
	chm.rwmutex.RLock()
	defer chm.rwmutex.RUnlock()
	if err := chm.checkStatus(); err != nil {
		return nil, err
	}
	return chm.errorCh, nil
}

//摘要方法
func (chm *ChannelManager) Summary() string {
	//模板
	chanmanSummaryTemplate := "status: %s, " +
		"requestChannel: %d/%d, " +
		"responseChannel: %d/%d, " +
		"entryChannel: %d/%d, " +
		"errorChannel: %d/%d"

	summary := fmt.Sprintf(chanmanSummaryTemplate,
		statusNameMap[chm.status],
		len(chm.reqCh), cap(chm.reqCh),
		len(chm.respCh), cap(chm.respCh),
		len(chm.entryCh), cap(chm.entryCh),
		len(chm.errorCh), cap(chm.errorCh))
	return summary
}

func (chm *ChannelManager) Status() ChannelManagerStatus {
	return chm.status
}
