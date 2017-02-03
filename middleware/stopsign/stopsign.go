package stopsign

/*
 * 停止信号
 */

//停止信号接口类型
type StopSignIntfs interface {
    // 置位停止信号。相当于发出停止信号。
    // 如果先前已发出过停止信号，那么该方法会返回false。
    Sign() bool
    // 判断停止信号是否已被发出。
    Signed() bool
    // 重置停止信号。相当于收回停止信号，并清除所有的停止信号处理记录。
    Reset()
    // 处理停止信号。
    // 参数code应该代表停止信号处理方的代号。该代号会出现在停止信号的处理记录中。
    Deal(code string)
    // 获取某一个停止信号处理方的处理计数。该处理计数会从相应的停止信号处理记录中获得。
    DealCount(code string) uint32
    // 获取停止信号被处理的总计数。
    DealTotal() uint32
    // 获取摘要信息。其中应该包含所有的停止信号处理记录。
    Summary() string
}