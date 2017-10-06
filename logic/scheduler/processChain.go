package scheduler

import (
    "fmt"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/log"
)

/*
 * 激活处理链
 * 一个独立的goroutine，循环从Entry通道中取出Entry
 * 然后交给独立的goroutine利用process chain去处理
 */
func (schdl *Scheduler) activateProcessChain() {
    go func() {
        schdl.processChain.SetFailFast(true)
        //对一个channel进行range操作，就是循环<-操作，并且在channel关闭之后能够自动结束
        for {
            entry, ok := schdl.getEntryChan().Get()
            if !ok {
                break //通道关闭
            }
            e, ok := entry.(basic.Entry)
            if !ok {
                continue
            }
            //每次从entry通道中取出一个entry，然后扔给一个独立的gorouting处理
            go schdl.processOneEntry(e)
        }
    }()
}

//将一个entry扔到processChain中去处理
func (schdl *Scheduler) processOneEntry(e basic.Entry) {
    defer func() {
        if p := recover(); p != nil {
            msg := fmt.Sprintf("Fatal entry Processing Error: %s\n", p)
            log.Warn(msg)
        }
    }()

    moudleCode := PROCESS_CHAIN_CODE
    //放入处理链，处理链上的节点自动处理，处理完毕就不必在理会了
    errs := schdl.processChain.SendAndProcess(e)
    if errs != nil {
        for _, err := range errs {
            schdl.sendError(err, moudleCode)
        }
    }
}