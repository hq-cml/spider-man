package scheduler

import (
	"github.com/hq-cml/spider-go/logic/processchain"
	chanman "github.com/hq-cml/spider-go/middleware/channel"
	"github.com/hq-cml/spider-go/middleware/requestcache"
	"github.com/hq-cml/spider-go/middleware/stopsign"
	"github.com/hq-cml/spider-go/middleware/pool"
	"sync"
	"github.com/hq-cml/spider-go/basic"
)

/*
 * 调度器
 */
// *Scheduler实现调度器的实现类型。
type Scheduler struct {
	grabMaxDepth   int                            // 爬取的最大深度。首次请求的深度为0。
	primaryDomain  string                         // 起始域名。
	channelManager *chanman.ChannelManager        // 通道管理器。
	poolManager    *pool.PoolManager              // Pool管理器。
	stopSign       *stopsign.StopSign             // 停止信号。
	processChain   *processchain.ProcessChain     // Item处理链条。
	requestCache   *requestcache.RequestCache     // Request缓存
	analyzeFuncs   []basic.AnalyzeResponseFunc    //处理器
	urlMap         sync.Map              		  // 已请求的URL的字典。
	urlCnt         uint64                         // sync.Map长度
	running        uint32                         // 运行标记。0表示未运行，1表示已运行，2表示已停止。

	downloaderCnt  uint64                         // 已启动的downloader协程数量
	analyzerCnt    uint64                         // 已启动的analyzer协程数量
}

// 调度器摘要信息的实现类型。
type SchedSummary struct {
	prefix              string // 前缀。
	running             uint32 // 运行标记。
	grabMaxDepth        int    // 爬取的最大深度。
	chanmanSummary      string // 通道管理器的摘要信息。
	reqCacheSummary     string // 请求缓存的摘要信息。
	poolmanSummary      string // pool管理器的摘要信息。
	processChainSummary string // 条目处理管道的摘要信息。
	urlCount            uint64 // 已请求的URL的计数。
	urlDetail           string // 已请求的URL的详细信息。
	stopSignSummary     string // 停止信号的摘要信息。

	downloaderCnt       uint64 // 已启动的downloader协程数量
	analyzerCnt         uint64 // 已启动的analyzer协程数量
}

// 组件的统一代号。
const (
	DOWNLOADER_CODE    = "downloader"
	ANALYZER_CODE      = "analyzer"
	PROCESS_CHAIN_CODE = "process_chain"
	SCHEDULER_CODE     = "scheduler"
	SUMMARY_CODE       = "summary"
)

const (
	RUNNING_STATUS_INIT    uint32 = 0
	RUNNING_STATUS_RUNNING        = 1
	RUNNING_STATUS_STOP           = 2
)

// 通道标志
const (
	CHANNEL_FLAG_REQUEST  = "request"
	CHANNEL_FLAG_RESPONSE = "response"
	CHANNEL_FLAG_ITEM     = "item"
	CHANNEL_FLAG_ERROR    = "error"
)