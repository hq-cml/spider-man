package scheduler

import (
    "net/http"
    "github.com/hq-cml/spider-go/logic/analyzer"
    "github.com/hq-cml/spider-go/logic/processchain"
    "fmt"
    "github.com/hq-cml/spider-go/helper/log"
    "errors"
    "sync/atomic"
    "github.com/hq-cml/spider-go/middleware/channelmanager"
    "github.com/hq-cml/spider-go/logic/downloader"
    "github.com/hq-cml/spider-go/middleware/stopsign"
    "time"
    "github.com/hq-cml/spider-go/helper/util"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/middleware/requestcache"
)


//New
func NewScheduler() SchedulerIntfs {
    return &Scheduler{}
}

/*
 * Scheduler实现SchedulerIntfs接口
 */
//Start开始
//TODO 重构
func (schdl *Scheduler)Start(channelLen uint, poolSize uint32, grabDepth uint32,
    httpClientGenerator GenHttpClientFunc,
    respAnalyzers []analyzer.AnalyzeResponseFunc,
    itemProcessors []processchain.ProcessItemFunc,
    firstHttpReq *http.Request) (err error) {

    //错误兜底
    defer func() {
        if e := recover(); e != nil {
            msg := fmt.Sprintf("Fatal Scheduler Error:%s\n", e)
            log.Warn(msg)
            err = errors.New(msg)
            return
        }
    }()

    //running状态设置！
    if atomic.LoadUint32(&schdl.running) == 1{
        err = errors.New("The scheduler has been started!\n") //已经开启，则退出，单例
        return
    }
    atomic.StoreUint32(&schdl.running, 1)

    //TODO 参数校验 & 赋值
    schdl.channelLen = channelLen
    schdl.poolSize = poolSize
    schdl.grabDepth = grabDepth

    //middleware生成
    schdl.channelManager = channelmanager.NewChannelManager(channelLen)

    if httpClientGenerator == nil {
        err = errors.New("The HTTP client generator list is invalid!\n") //已经开启，则退出，单例
        return
    }

    if schdl.downloaderPool, err = downloader.NewDownloaderPool(poolSize,
        func() downloader.DownloaderIntfs{
            return downloader.NewDownloader(httpClientGenerator())
        },
    ); err != nil{
        err = errors.New(fmt.Sprintf("Occur error when gen downloader pool: %s\n", err))
        return
    }

    if schdl.analyzerPool, err = analyzer.NewAnalyzerPool(poolSize,
        func() analyzer.AnalyzerIntfs{
            return analyzer.NewAnalyzer()
        },
    ); err != nil {
        err = errors.New(fmt.Sprintf("Occur error when gen downloader pool: %s\n", err))
        return
    }

    if itemProcessors == nil {
        return errors.New("The item processor list is invalid!")
    }
    for i, ip := range itemProcessors {
        if ip == nil {
            return errors.New(fmt.Sprintf("The %dth item processor is invalid!", i))
        }
    }
    schdl.processChain = processchain.NewProcessChain(itemProcessors)

    if schdl.stopSign == nil {
        schdl.stopSign = stopsign.NewStopSign()
    } else {
        schdl.stopSign.Reset()
    }

    schdl.requestCache = requestcache.NewRequestCache()
    schdl.urlMap = make(map[string]bool)

    schdl.activeDownloaders()
    schdl.activateAnalyzers(respAnalyzers)
    schdl.openItemPipeline()
    schdl.schedule(10 * time.Millisecond)

    if firstHttpReq == nil {
        return errors.New("The first HTTP request is invalid!")
    }
    if schdl.primaryDomain, err = util.GetPrimaryDomain(firstHttpReq.Host); err != nil {
        return err
    }

    firstReq := basic.NewRequest(firstHttpReq, 0) //深度0
    schdl.requestCache.Put(firstReq)

    return nil

}

/*
 * 激活分析器，开始分析，分析工作由异步的goroutine进行负责
 * 无限循环，从响应通道中获取响应，完成分析工作
 */
func (schdl *Scheduler) activateAnalyzers(respAnalyzers []analyzer.AnalyzeResponseFunc) {
    go func() {
       for { //无限循环
           response, ok := <- schdl.getResponseChan()
           if !ok {
               //通道已关闭
               break
           }
           //启动异步分析
           go schdl.analyze(respAnalyzers, response)
       }
    }()
}

//实际分析工作
func (schdl *Scheduler) analyze(respAnalyzers []analyzer.AnalyzeResponseFunc, response basic.Response) {
    defer func() {
       if p := recover(); p!=nil{
           msg := fmt.Sprintf("Fatal Analysis Error: %s\n", p)
           log.Warn(msg)
       }
    }()

    analyzer, err := schdl.analyzerPool.Get()
    if err != nil {
        msg := fmt.Sprintf("Analyzer pool error: %s", err)
        schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        return
    }
    defer func(){ //注册延时归还
        err = schdl.analyzerPool.Put(analyzer)
        if err != nil {
            msg := fmt.Sprintf("Analyzer pool error: %s", err)
            schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        }
    }()

    moudleCode := generateModuleCode(ANALYZER_CODE, analyzer.Id())
    dataList, errs := analyzer.Analyze(respAnalyzers, response)
    if errs != nil {
        for _, err := range errs {
            schdl.sendError(err, moudleCode)
        }

    }
    //Analyze返回值是一个列表，其中元素可能是两种类型：请求 or 条目
    if dataList != nil {
        for _, data := range dataList {
            if data == nil {
                continue
            }

            switch  d:= data.(type) {
            case *basic.Request:
                schdl.sendRequestToCache(*d, moudleCode)
            case *basic.Item:
                schdl.sendItem(*d, moudleCode)
            default:
                msg := fmt.Sprintf("Unsported data type:%T! (value=%v)\n", d, d)
                schdl.sendError(errors.New(msg), moudleCode)
            }
        }
        schdl.sendResponse(*response, moudleCode)
    }
}

/*
 * 激活下载器，开始下载，下载工作由异步的goroutine进行负责
 * 无限循环，从请求通道中获取请求，完成下载任务
 */
func (schdl *Scheduler) activeDownloaders() {
    go func() {
        //无限循环，从请求通道中获取请求
        for {
            request, ok := <- schdl.getReqestChan()
            if !ok {
                //通道已关闭
                break
            }
            //每个请求都交给一个独立的goroutine来处理
            go schdl.download(request)
        }
    }()
}


//实际下载
func (schdl *Scheduler) download(request basic.Request) {
    defer func() {
       if p := recover(); p != nil {
           msg := fmt.Sprintf("Fatal Download Error: %s\n", p)
           log.Warn(msg)
       }
    }()

    downloader, err := schdl.downloaderPool.Get()
    if err != nil {
        msg := fmt.Sprintf("Downloader pool error: %s", err)
        schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        return
    }
    defer func(){ //注册延时归还
        err = schdl.downloaderPool.Put(downloader)
        if err != nil {
            msg := fmt.Sprintf("Downloader pool error: %s", err)
            schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        }
    }()

    moudleCode := generateModuleCode(DOWNLOADER_CODE, downloader.Id())
    response, err := downloader.Download(request)
    if err != nil {
        schdl.sendError(err, moudleCode)
    }
    if response != nil {
        schdl.sendResponse(*response, moudleCode)
    }
}