package processchain

import (
	"errors"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
	"sync/atomic"
)

/*
 * Entry处理处理链，每个entry都会被处理链进行流式处理
 * 具体的处理逻辑就是这些链中的每个函数，交由用户自定制
 */

//New, 创建处理链
func NewProcessChain(entryProcessors []ProcessEntryFunc) ProcessChainIntfs {
	if entryProcessors == nil {
		panic(errors.New("Invalid entry processor list!"))
	}

	pc := make([]ProcessEntryFunc, 0)

	for k, v := range entryProcessors {
		if v == nil {
			panic(errors.New(fmt.Sprint("Invalid entry processor:", k)))
		}
		pc = append(pc, v)
	}

	return &ProcessChain{
		entryProcessors: pc,
	}
}

//*ProcessChain实现接口ProcessChainIntfs
//向处理链发送entry
func (pc *ProcessChain) Send(entry basic.Entry) []error {
	atomic.AddUint64(&pc.processingNumber, 1)                //原子加1
	defer atomic.AddUint64(&pc.processingNumber, ^uint64(0)) //原子减1
	atomic.AddUint64(&pc.sent, 1)

	errs := []error{}
	if entry == nil {
		errs = append(errs, errors.New("entry is invalid!"))
	}

	atomic.AddUint64(&pc.accepted, 1)
	var currentEntry basic.Entry = entry //备份出一份本地entry
	//链式处理
	for _, processFunc := range pc.entryProcessors {
		processedEntry, err := processFunc(currentEntry)

		if err != nil {
			errs = append(errs, err)
			if pc.failFast {
				break
			}
		}

		if processedEntry != nil {
			currentEntry = processedEntry
		}
	}

	atomic.AddUint64(&pc.processed, 1)
	return errs
}

func (pc *ProcessChain) FailFast() bool {
	return pc.failFast
}

//设置是否快速失败。
func (pc *ProcessChain) SetFailFast(failFast bool) {
	pc.failFast = failFast
}

// 获得已发送、已接受和已处理的条目的计数值。
func (pc *ProcessChain) Count() []uint64 {
	counts := make([]uint64, 3)
	counts[0] = atomic.LoadUint64(&pc.sent)
	counts[1] = atomic.LoadUint64(&pc.accepted)
	counts[2] = atomic.LoadUint64(&pc.processed)
	return counts
}

// 获取正在被处理的条目的数量。
func (pc *ProcessChain) ProcessingNumber() uint64 {
	return atomic.LoadUint64(&pc.processingNumber)
}

//获取摘要信息
func (pc *ProcessChain) Summary() string {
	var summaryTemplate = "failFast: %v, processorNumber: %d," +
		" sent: %d, accepted: %d, processed: %d, processingNumber: %d"

	counts := pc.Count()
	summary := fmt.Sprintf(summaryTemplate,
		pc.failFast, len(pc.entryProcessors), counts[0], counts[1], counts[2], pc.ProcessingNumber())
	return summary
}
