package scheduler

import (
    "fmt"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/log"
)

/*
 * 激活Item处理器链
 * 一个独立的goroutine，循环从Item通道中取出Item
 * 然后交给独立的goroutine利用process chain去处理
 */
func (schdl *Scheduler) activateItemProcessor() {
    go func() {
        schdl.processChain.SetFailFast(true)
        //对一个channel进行range操作，就是循环<-操作，并且在channel关闭之后能够自动结束
        for {
            item, ok := schdl.getItemChan().Get()
            if !ok {
                break //通道关闭
            }
            e, ok := item.(basic.Item)
            if !ok {
                continue
            }
            //每次从item通道中取出一个item，然后启动独立的gorouting异步处理
            go schdl.processOneItem(e)
        }
    }()
}

//将一个item扔到processChain中去处理
func (schdl *Scheduler) processOneItem(e basic.Item) {
    defer func() {
        if p := recover(); p != nil {
            msg := fmt.Sprintf("Fatal item Processing Error: %s\n", p)
            log.Err(msg)
        }
    }()

    moudleCode := PROCESS_CHAIN_CODE
    //放入处理链，处理链上的节点自动处理，处理完毕就不必在理会了
    errs := schdl.processChain.DoProcess(e)
    if errs != nil {
        for _, err := range errs {
            schdl.sendError(err, moudleCode)
        }
    }
}