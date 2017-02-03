package scheduler

import (
    "net/http"
    "github.com/hq-cml/spider-go/logic/analyzer"
    "github.com/hq-cml/spider-go/logic/itempipeline"
)

/*
 * 调度控制器
 */

//用来生成Http客户端的函数的类型
type GenHttpClient func() *http.Client

//调度器接口类型
type SchedulerIntfs interface{
    // 开启调度器。
    // 调用该方法会使调度器创建和初始化各个组件。在此之后，调度器会激活爬取流程的执行。
    // 参数channelArgs代表通道参数的容器。
    // 参数poolBaseArgs代表池基本参数的容器。
    // 参数crawlDepth代表了需要被爬取的网页的最大深度值。深度大于此值的网页会被忽略。
    // 参数httpClientGenerator代表的是被用来生成HTTP客户端的函数。
    // 参数respParsers的值应为分析器所需的被用来解析HTTP响应的函数的序列。
    // 参数itemProcessors的值应为需要被置入条目处理管道中的条目处理器的序列。
    // 参数firstHttpReq即代表首次请求。调度器会以此为起始点开始执行爬取流程。
    Start(channelLen uint,
        poolSize uint32,
        grabDepth uint32,
        httpClientGenerator GenHttpClient,
        respParsers []analyzer.ParseResponse,
        itemProcessors []itempipeline.ProcessItem,
        firstHttpReq *http.Request) (err error)
    //调用该方法会停止调度器的运行。所有处理模块执行的流程都会被中止。
    Stop() bool
    // 判断调度器是否正在运行。
    Running() bool
    // 获得错误通道。调度器以及各个处理模块运行过程中出现的所有错误都会被发送到该通道。
    // 若该方法的结果值为nil，则说明错误通道不可用或调度器已被停止。
    ErrorChan() <-chan error
    // 判断所有处理模块是否都处于空闲状态。
    Idle() bool
    // 获取摘要信息。
    Summary(prefix string) SchedSummaryIntfs
}
