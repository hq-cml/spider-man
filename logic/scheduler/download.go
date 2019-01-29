package scheduler

import (
    "fmt"
    "errors"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/log"
    "github.com/hq-cml/spider-go/middleware/pool"
    "github.com/hq-cml/spider-go/logic/downloader"
)

/*
 * 激活下载器，开始下载，下载器是一个独立的的goroutine无限循环，从请求通道中获取请求
 * 每个请求都扔给独立的goroutine去处理，但是goroutine并不一定能够立刻开始下载工作
 * 同时能够执行下载的goroutine数量,受到下载器池容量的制约, 其他goroutine会阻塞
 */
func (schdl *Scheduler) activateDownloaders() {
    //downloader是异步独立的goroutine
    go func() {
        //无限循环，从请求通道中获取请求，每个请求都扔给独立的goroutine去处理
        for {
            request, ok := schdl.getReqestChan().Get()
            if !ok {
                //通道已关闭
                break
            }
            //类型断言
            req, ok := request.(basic.Request)
            if !ok {
                continue
            }
            //每个请求都交给一个独立的goroutine来处理
            go schdl.download(req)
        }
    }()
}

/*
 * 实际下载工作，下载goroutine的逻辑
 * 但是全部下载goroutine是受到下载器池子的约束的
 */
func (schdl *Scheduler) download(request basic.Request) {
    //panic错误兜底
    defer func() {
        if p := recover(); p != nil {
            msg := fmt.Sprintf("Fatal Download Error: %s\n", p)
            log.Warn(msg)
        }
    }()

    //下载器池中取票
    entity, err := schdl.getDownloaderPool().Get()
    if err != nil {
        msg := fmt.Sprintf("Downloader pool error: %s", err)
        schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        return
    }
    defer func() { //注册延时归还
        err = schdl.getDownloaderPool().Put(entity)
        if err != nil {
            msg := fmt.Sprintf("Downloader pool error: %s", err)
            schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        }
    }()

    //断言转换
    dl, ok := entity.(*downloader.Downloader)
    if !ok {
        msg := fmt.Sprintf("Downloader pool Wrong type")
        schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        return
    }
    moudleCode := generateModuleCode(DOWNLOADER_CODE, dl.Id())
    response, err := dl.Download(request)
    if err != nil {
        schdl.sendError(err, moudleCode)
    }
    if response != nil {
        schdl.sendToRespChan(*response, moudleCode)
    }
}

//获取Pool管理器持有的下载器Pool。
func (schdl *Scheduler) getDownloaderPool() pool.PoolIntfs {
    p, err := schdl.poolManager.GetPool("downloader")
    if err != nil {
        panic(err)
    }
    return p
}

//发送响应到通道管理器中的响应通道
func (schdl *Scheduler) sendToRespChan(resp basic.Response, mouduleCode string) bool {
    if schdl.stopSign.Signed() {
        //如果stop标记已经生效，则通道管理器可能已经关闭，此时不应该再进行通道写入
        schdl.stopSign.Deal(mouduleCode)
        return false
    }

    schdl.getResponseChan().Put(resp)
    return true
}