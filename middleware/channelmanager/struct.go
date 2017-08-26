package channelmanager

import (
	"github.com/hq-cml/spider-go/basic"
	"sync"
	"errors"
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
var statusNameMap = map[ChannelManagerStatus]string{
	CHANNEL_MANAGER_STATUS_UNINITIALIZED: "uninitialized",
	CHANNEL_MANAGER_STATUS_INITIALIZED:   "inititalized",
	CHANNEL_MANAGER_STATUS_CLOSED:        "closed",
}


//channel管理器实现类型
type ChannelManager struct {
	channelParams ChannelParams  //通道长度
	reqCh         chan basic.Request   //请求通道
	respCh        chan basic.Response  //响应通道
	entryCh       chan basic.Entry     //entry通道
	errorCh       chan error           //错误通道
	status        ChannelManagerStatus //channel管理器状态
	rwmutex       sync.RWMutex         //读写锁
}

//通道管理器实现类型
type NChannelManager struct {
	channel       map[string]SpiderChannelIntfs
	status        ChannelManagerStatus //channel管理器状态
	rwmutex       sync.RWMutex         //读写锁
}

//通道参数的容器。
type ChannelParams struct {
	reqChanLen   uint   // 请求通道的长度。
	respChanLen  uint   // 响应通道的长度。
	entryChanLen uint   // 条目通道的长度。
	errorChanLen uint   // 错误通道的长度。
	description  string // 描述。
}