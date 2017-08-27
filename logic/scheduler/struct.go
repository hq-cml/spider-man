package scheduler

import (
	"github.com/hq-cml/spider-go/logic/analyzer"
	"github.com/hq-cml/spider-go/logic/downloader"
	"github.com/hq-cml/spider-go/logic/processchain"
	"github.com/hq-cml/spider-go/middleware/channelmanager"
	"github.com/hq-cml/spider-go/middleware/requestcache"
	"github.com/hq-cml/spider-go/middleware/stopsign"
)

/*
 * 调度控制器
 */
// *Scheduler实现调度器的实现类型。
type Scheduler struct {
	grabDepth      uint32                             // 爬取的最大深度。首次请求的深度为0。
	primaryDomain  string                             // 主域名。
	channelManager *channelmanager.ChannelManager // 通道管理器。
	stopSign       stopsign.StopSignIntfs             // 停止信号。
	downloaderPool downloader.DownloaderPoolIntfs     // 网页下载器池。
	analyzerPool   analyzer.AnalyzerPoolIntfs         // 分析器池。
	processChain   processchain.ProcessChainIntfs     // 条目处理管道。
	requestCache   *requestcache.RequestCache     // 请求缓存。
	urlMap         map[string]bool                    // 已请求的URL的字典。
	running        uint32                             // 运行标记。0表示未运行，1表示已运行，2表示已停止。
}

// 调度器摘要信息的实现类型。
type SchedSummary struct {
	prefix              string              // 前缀。
	running             uint32              // 运行标记。
	grabDepth           uint32              // 爬取的最大深度。
	chanmanSummary      string              // 通道管理器的摘要信息。
	reqCacheSummary     string              // 请求缓存的摘要信息。
	dlPoolLen           uint32              // 网页下载器池的长度。
	dlPoolCap           uint32              // 网页下载器池的容量。
	analyzerPoolLen     uint32              // 分析器池的长度。
	analyzerPoolCap     uint32              // 分析器池的容量。
	processChainSummary string              // 条目处理管道的摘要信息。
	urlCount            int                 // 已请求的URL的计数。
	urlDetail           string              // 已请求的URL的详细信息。
	stopSignSummary     string              // 停止信号的摘要信息。
}

// 组件的统一代号。
const (
	DOWNLOADER_CODE    = "downloader"
	ANALYZER_CODE      = "analyzer"
	PROCESS_CHAIN_CODE = "process_chain"
	SCHEDULER_CODE     = "scheduler"
)

const (
	RUNNING_STATUS_INIT    uint32 = 0
	RUNNING_STATUS_RUNNING        = 1
	RUNNING_STATUS_STOP           = 2
)
