package processchain

import (
	"github.com/hq-cml/spider-go/basic"
)

/*
 * entry处理处理链，每一个节点都是一个处理函数
 * 所有被分析出的entry，需要经过处理链的逐个处理
 */

type ProcessChainIntfs interface {
	//向处理链发送entry
	Send(entry basic.Entry) []error
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

// 条目处理管道的实现类型。
type ProcessChain struct {
	entryProcessors  []basic.ProcessEntryFunc // 条目处理器的列表。
	failFast         bool                     // 表示处理是否需要快速失败的标志位。
	sent             uint64                   // 已被发送的条目的数量。
	accepted         uint64                   // 已被接受的条目的数量。
	processed        uint64                   // 已被处理的条目的数量。
	processingNumber uint64                   // 正在被处理的条目的数量。
}
