package processchain

import (
	"errors"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
	"sync/atomic"
)

//TODO 重命名processChain
/*
 * Item处理处理链，每个item都会被处理链进行流式处理
 * 具体的处理逻辑就是这些链中的每个函数，交由用户自定制
 */
// 条目处理链类型。
type ProcessChain struct {
	itemProcessors   []basic.ProcessItemFunc // 条目处理器的列表。
	failFast         bool                    // 表示处理是否需要快速失败的标志位。
	sent             uint64                  // 已被发送的条目的数量。
	accepted         uint64                  // 已被接受的条目的数量。
	processed        uint64                  // 已被处理的条目的数量。
	processingNumber uint64                  // 正在被处理的条目的数量。
}

//New, 创建处理链
func NewProcessChain(itemProcessors []basic.ProcessItemFunc) *ProcessChain {
	//用户自定制处理链，如果是空的，则程序无法正常运转
	if itemProcessors == nil {
		panic(errors.New("Invalid item processor list!"))
	}

	pc := make([]basic.ProcessItemFunc, 0)

	for k, v := range itemProcessors {
		if v == nil {
			panic(errors.New(fmt.Sprint("Invalid item processor:", k)))
		}
		pc = append(pc, v)
	}

	return &ProcessChain {
		itemProcessors: pc,
	}
}

//向处理链发送item，调用处理链自动进行处理
func (pc *ProcessChain) SendAndProcess(item basic.Item) []error {
	atomic.AddUint64(&pc.processingNumber, 1)                //原子加1
	defer atomic.AddUint64(&pc.processingNumber, ^uint64(0)) //原子减1
	atomic.AddUint64(&pc.sent, 1)

	errs := []error{}
	if item == nil {
		errs = append(errs, errors.New("item is invalid!"))
		return errs
	}

	atomic.AddUint64(&pc.accepted, 1)
	var currentItem basic.Item = item //备份出一份本地item，其实没啥用，map是引用类型
	//链式处理
	for _, processFunc := range pc.itemProcessors {
		processedItem, err := processFunc(currentItem)

		if err != nil {
			errs = append(errs, err)
			if pc.failFast {
				break
			}
		}

		if processedItem != nil {
			currentItem = processedItem
		}
	}

	atomic.AddUint64(&pc.processed, 1)
	return errs
}

//该值表示当前的条目处理链是否是快速失败的。即只要对某个条目的处理流程在某一个步骤上出错，
//那么处理链就会忽略掉后续的所有处理步骤并报告错误。
func (pc *ProcessChain) FailFast() bool {
	return pc.failFast
}

//设置是否快速失败。
func (pc *ProcessChain) SetFailFast(failFast bool) {
	pc.failFast = failFast
}

//获取正在被处理的条目的数量。
func (pc *ProcessChain) ProcessingNumber() uint64 {
	return atomic.LoadUint64(&pc.processingNumber)
}

//获取摘要信息
func (pc *ProcessChain) Summary(prefix string) string {
	sent := atomic.LoadUint64(&pc.sent)				//已发送
	accepted := atomic.LoadUint64(&pc.accepted)		//已接收
	processed := atomic.LoadUint64(&pc.processed)	//已处理

	summary := fmt.Sprintf(prefix + "FailFast: %v, processorNumber: %d\n" + prefix+ "sent: %d, accepted: %d, processed: %d, processingNumber: %d\n",
		pc.failFast, len(pc.itemProcessors), sent, accepted, processed, pc.ProcessingNumber())
	return summary
}
