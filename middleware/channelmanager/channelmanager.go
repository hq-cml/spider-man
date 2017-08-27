package channelmanager

import (
	"errors"
	"bytes"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
)

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
