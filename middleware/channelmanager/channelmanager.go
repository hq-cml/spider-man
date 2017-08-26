package channelmanager

import (
	"errors"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
)

//New
func NewChannelManager(chp ChannelParams) ChannelManager {
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

// 初始化通道管理器。
// 参数channelArgs代表通道参数的容器。
// 参数reset指明是否重新初始化通道管理器。
func (chm *ChannelManager) Init(chp ChannelParams, reset bool) bool {
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

// 关闭通道管理器
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


//创建通道参数的容器。
func NewChannelParams(reqChanLen uint, respChanLen uint, entryChanLen uint, errorChanLen uint) ChannelParams {
	return ChannelParams{
		reqChanLen:   reqChanLen,
		respChanLen:  respChanLen,
		entryChanLen: entryChanLen,
		errorChanLen: errorChanLen,
	}
}

func (p *ChannelParams) Check() error {
	if p.reqChanLen == 0 {
		return errors.New("The request channel max length (capacity) can not be 0!\n")
	}
	if p.respChanLen == 0 {
		return errors.New("The response channel max length (capacity) can not be 0!\n")
	}
	if p.entryChanLen == 0 {
		return errors.New("The entry channel max length (capacity) can not be 0!\n")
	}
	if p.errorChanLen == 0 {
		return errors.New("The error channel max length (capacity) can not be 0!\n")
	}
	return nil
}

//通道参数的容器的描述模板。
func (args *ChannelParams) String() string {
	if args.description == "" {
		args.description = fmt.Sprintf("{ reqChanLen: %d, respChanLen: %d, entryChanLen: %d, errorChanLen: %d }", args.reqChanLen, args.respChanLen,
			args.entryChanLen, args.errorChanLen)
	}
	return args.description
}

// 获得请求通道的长度。
func (p *ChannelParams) ReqChanLen() uint {
	return p.reqChanLen
}

// 获得响应通道的长度。
func (p *ChannelParams) RespChanLen() uint {
	return p.respChanLen
}

// 获得条目通道的长度。
func (p *ChannelParams) EntryChanLen() uint {
	return p.entryChanLen
}

// 获得错误通道的长度。
func (p *ChannelParams) ErrorChanLen() uint {
	return p.errorChanLen
}