package channelmanager
/*
 * channel管理器
 * 框架中有四种类型的数据需要管道传递：请求、响应、条目、Error
 */
import (
	"errors"
	"bytes"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
	"sync"
)

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

//通道管理器实现类型
type ChannelManager struct {
	channel       map[string]basic.SpiderChannelIntfs
	status        ChannelManagerStatus //channel管理器状态
	rwmutex       sync.RWMutex         //读写锁
}

//New
func NewChannelManager() *ChannelManager {
	chm := &ChannelManager{
		status:CHANNEL_MANAGER_STATUS_INITIALIZED,
		channel:make(map[string]basic.SpiderChannelIntfs),
	}
	return chm
}

//关闭通道管理器
func (chm *ChannelManager) Close() bool {
	//写锁保护
	chm.rwmutex.Lock()
	defer chm.rwmutex.Unlock()

	if chm.status != CHANNEL_MANAGER_STATUS_INITIALIZED {
		return false
	}

	//逐个关闭
	for _, c := range chm.channel {
		c.Close()
	}
	chm.status = CHANNEL_MANAGER_STATUS_CLOSED

	return true
}

//注册一个新的通道进入管理器
func (chm *ChannelManager) RegisterOneChannel(name string, c basic.SpiderChannelIntfs) error {
	//写锁保护
	chm.rwmutex.Lock()
	defer chm.rwmutex.Unlock()

	if _, ok := chm.channel[name]; ok {
		return errors.New("Already Exist channel")
	}
	chm.channel[name] = c

	return nil
}

//获取request通道
func (chm *ChannelManager) GetOneChannel(name string) (basic.SpiderChannelIntfs, error) {
	//读锁保护
	chm.rwmutex.RLock()
	defer chm.rwmutex.RUnlock()

	c, ok := chm.channel[name]
	if !ok {
		return nil, errors.New("Not found")
	}

	return c, nil
}

//摘要方法
func (chm *ChannelManager) Summary() string {
	//读锁保护
	chm.rwmutex.RLock()
	defer chm.rwmutex.RUnlock()

	var buff bytes.Buffer
	buff.WriteString("ChannelManager Status:"+statusNameMap[chm.status]+"\n")
	for k, c := range chm.channel {
		buff.WriteString(fmt.Sprint("%s: Len:%d, Cap:/%d\n ", k, c.Len(), c.Cap()))
	}

	return buff.String()
}

func (chm *ChannelManager) Status() ChannelManagerStatus {
	return chm.status
}
