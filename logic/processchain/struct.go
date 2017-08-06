package processchain

import "github.com/hq-cml/spider-go/basic"

/*
 * Item处理处理链，每一个节点都是一个处理函数
 * 所有被分析出的item，需要经过处理链的逐个处理
 */

// 被用来处理Item的函数的类型
type ProcessItemFunc func(item basic.Item) (result basic.Item, err error)

type ProcessChainIntfs interface {
    //向处理链发送Item
    Send(item basic.Item) []error
    // 该值表示当前的条目处理链是否是快速失败的。即只要对某个条目的处理流程在某一个步骤上出错，
    // 那么处理链就会忽略掉后续的所有处理步骤并报告错误。
    FailFast() bool
    // 设置是否快速失败。
    SetFailFast(failFast bool)
    // 获得已发送、已接受和已处理的条目的计数值。
    // 更确切地说，作为结果值的切片总会有三个元素值。这三个值会分别代表前述的三个计数。
    Count() []uint64
    // 获取正在被处理的条目的数量。
    ProcessingNumber() uint64
    //获取摘要信息
    Summary() string
}