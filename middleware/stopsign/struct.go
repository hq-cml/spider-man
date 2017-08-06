package stopsign

import (
    "sync"
)

/*
 * 停止信号
 * 实现"一方发送，多方接收的效果"
 */

//停止信号接口类型
type StopSignIntfs interface {
    Sign() bool //发出停止信号。如果先前已发出过停止信号，那么该方法会返回false。
    Signed() bool// 判断停止信号是否已被发出。
    Reset()// 重置停止信号。相当于收回停止信号，并清除所有的停止信号处理记录。
    Deal(code string)// 处理停止信号。当处理了停止信号之后，停止信号的处理方应该调用停止信号Deal方法，表示已经对该信号处理完毕。参数code应该代表停止信号处理方的代号。
    DealCount(code string) uint32// 获取某一个停止信号处理方的处理计数。
    DealTotal() uint32// 获取停止信号被处理的总计数。
    Summary() string// 获取摘要信息。其中应该包含所有的停止信号处理记录。
}

//停止信号实现
type StopSign struct {
    rwmutex   sync.RWMutex         //保护锁
    signed    bool                 //信号是否已发送标志
    dealCountMap map[string]uint32 //处理方处理计数map
}
