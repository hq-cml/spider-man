package scheduler

import (
    "fmt"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/log"
)

//从错误通道中接收和报告错误。
func (schdl *Scheduler)activateProcessError() {
    go func() {
        defer func() {
            if p := recover(); p != nil {
                msg := fmt.Sprintf("ProcessError panic: %s\n", p)
                log.Warn(msg)
            }
        }()

        for {
            err, ok := schdl.getErrorChan().Get()
            if !ok {
                return
            }
            e, ok := err.(*basic.SpiderError)
            if !ok {
                continue
            }

            errMsg := fmt.Sprintf("Received from error channel:ErrType: %s, Detail: %s", e.Type(), e.Error())
            log.Warn(errMsg)
        }
    }()
}
