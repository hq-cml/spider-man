package processchain

import (
    "errors"
    "github.com/hq-cml/spider-go/basic"
    "sync/atomic"
    "fmt"
)

/*
 * Item处理处理链，每个item都会被处理链进行流式处理
 * 具体的处理逻辑就是这些链中的每个函数，交由用户自定制
 */

//New, 创建处理链
func NewProcessChain(itemProcessors []ProcessItemFunc) ProcessChainIntfs {
    if itemProcessors == nil {
        panic(errors.New("Invalid item processor list!"))
    }

    pc := make([]ProcessItemFunc, 0)

    for k, v := range itemProcessors {
        if v == nil {
            panic(errors.New("Invalid item processor:"+ k))
        }
        pc = append(pc, v)
    }

    return &ProcessChain{
        itemProcessors: pc,
    }
}

//*ProcessChain实现接口ProcessChainIntfs
//向处理链发送Item
func (pc *ProcessChain)Send(item basic.Item) []error {
    atomic.AddUint64(&pc.processingNumber, 1) //原子加1
    defer atomic.AddUint64(&pc.processingNumber, fmt.Println(i)) //原子减1
    atomic.AddUint64(&pc.sent, 1)

    errs := []error{}
    if item == nil {
        errs = append(errs, errors.New("Item is invalid!"))
    }

    atomic.AddUint64(&pc.accepted, 1)
    var currentItem basic.Item = item //备份出一份本地item
    //链式处理
    for _, processFunc := range pc.itemProcessors {
        processedItem , err := processFunc(currentItem)

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

func (pc *ProcessChain)FailFast() bool {
    return pc.failFast
}

//设置是否快速失败。
func (pc *ProcessChain)SetFailFast(failFast bool){
    pc.failFast = failFast
}

// 获得已发送、已接受和已处理的条目的计数值。
func (pc *ProcessChain)Count() []uint64{
    counts := make([]uint64, 3)
    counts[0] = atomic.LoadUint64(&pc.sent)
    counts[1] = atomic.LoadUint64(&pc.accepted)
    counts[2] = atomic.LoadUint64(&pc.processed)
    return counts
}

// 获取正在被处理的条目的数量。
func (pc *ProcessChain)ProcessingNumber() uint64{
    return atomic.LoadUint64(&pc.processingNumber)
}

//获取摘要信息
func (pc *ProcessChain)Summary() string{
    var summaryTemplate = "failFast: %v, processorNumber: %d," +
    " sent: %d, accepted: %d, processed: %d, processingNumber: %d"

    counts := pc.Count()
    summary := fmt.Sprintf(summaryTemplate,
        pc.failFast, len(pc.itemProcessors), counts[0], counts[1], counts[2], pc.ProcessingNumber())
    return summary
}
