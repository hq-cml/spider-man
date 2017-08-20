package scheduler

import (
    "net/http"
    "github.com/hq-cml/spider-go/logic/analyzer"
    "github.com/hq-cml/spider-go/logic/processchain"
    "github.com/hq-cml/spider-go/middleware/stopsign"
    "github.com/hq-cml/spider-go/middleware/channelmanager"
    "github.com/hq-cml/spider-go/middleware/requestcache"
    "github.com/hq-cml/spider-go/logic/downloader"
    "github.com/hq-cml/spider-go/basic"
)

/*
 * 调度控制器
 */
//用来生成Http客户端的函数的类型
type GenHttpClientFunc func() *http.Client

//调度器接口类型
type SchedulerIntfs interface{
    // 开启调度器。
    // 调用该方法会使调度器创建和初始化各个组件。在此之后，调度器会激活爬取流程的执行。
    // 参数channelParams代表通道参数的容器。
    // 参数poolParams代表池基本参数的容器。
    // 参数crawlDepth代表了需要被爬取的网页的最大深度值。深度大于此值的网页会被忽略。
    // 参数httpClientGenerator代表的是被用来生成HTTP客户端的函数。
    // 参数respParsers的值应为分析器所需的被用来解析HTTP响应的函数的序列。
    // 参数entryProcessors的值应为需要被置入条目处理管道中的条目处理器的序列。
    // 参数firstHttpReq即代表首次请求。调度器会以此为起始点开始执行爬取流程。
    Start(channelParams basic.ChannelParams, poolParams basic.PoolParams,
        grabDepth uint32,
        httpClientGenerator GenHttpClientFunc,
        respParsers []analyzer.AnalyzeResponseFunc,
        entryProcessors []processchain.ProcessEntryFunc,
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

// 调度器摘要信息的接口类型。
type SchedSummaryIntfs interface {
    String() string               // 获得摘要信息的一般表示。
    Detail() string               // 获取摘要信息的详细表示。
    Same(other SchedSummaryIntfs) bool // 判断是否与另一份摘要信息相同。
}

// *Scheduler实现调度器的实现类型。
type Scheduler struct {
    channelParams   basic.ChannelParams   // 通道参数的容器。
    poolParams      basic.PoolParams   // 池基本参数的容器。
    grabDepth      uint32                             // 爬取的最大深度。首次请求的深度为0。
    primaryDomain  string                             // 主域名。
    channelManager channelmanager.ChannelManagerIntfs // 通道管理器。
    stopSign       stopsign.StopSignIntfs             // 停止信号。
    downloaderPool downloader.DownloaderPoolIntfs     // 网页下载器池。
    analyzerPool   analyzer.AnalyzerPoolIntfs         // 分析器池。
    processChain   processchain.ProcessChainIntfs     // 条目处理管道。
    requestCache   requestcache.RequestCacheIntfs     // 请求缓存。
    urlMap         map[string]bool                    // 已请求的URL的字典。
    running        uint32                             // 运行标记。0表示未运行，1表示已运行，2表示已停止。
}

// 调度器摘要信息的实现类型。
type SchedSummary struct {
    prefix              string            // 前缀。
    running             uint32            // 运行标记。
    channelParams   basic.ChannelParams   // 通道参数的容器。
    poolParams      basic.PoolParams   // 池基本参数的容器。
    grabDepth          uint32            // 爬取的最大深度。
    chanmanSummary      string            // 通道管理器的摘要信息。
    reqCacheSummary     string            // 请求缓存的摘要信息。
    dlPoolLen           uint32            // 网页下载器池的长度。
    dlPoolCap           uint32            // 网页下载器池的容量。
    analyzerPoolLen     uint32            // 分析器池的长度。
    analyzerPoolCap     uint32            // 分析器池的容量。
    processChainSummary string            // 条目处理管道的摘要信息。
    urlCount            int               // 已请求的URL的计数。
    urlDetail           string            // 已请求的URL的详细信息。
    stopSignSummary     string            // 停止信号的摘要信息。
}

// 组件的统一代号。
const (
    DOWNLOADER_CODE   = "downloader"
    ANALYZER_CODE     = "analyzer"
    PROCESS_CHAIN_CODE = "process_chain"
    SCHEDULER_CODE    = "scheduler"
)

const (
    RUNNING_STATUS_INIT uint32 = 0
    RUNNING_STATUS_RUNNING  = 1
    RUNNING_STATUS_STOP     = 2
)
