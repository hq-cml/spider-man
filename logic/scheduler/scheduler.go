package scheduler

/*
 * 调度器
 * 框架的核心，将所有的中间件和逻辑组件进行整合、同步、协调，组装成爬虫的核心逻辑
 *
 * 调度的最核心的一步是对于request缓冲区的利用，而不是直接利用request通道：
 * 整个框架最有可能阻塞的是request通道，因为无法预知分析出的页面会产出多少新的request
 * 如果request通道被打满阻塞，可能会导致整个框架的阻塞，所以利用request缓冲区来避免
 * 每一轮都会先计算出request通道的剩余容量，然后从缓冲中取出相同的数量的请求放入通道
 * 这样就可以完全防止request通道的阻塞，从而保证所有通道的顺畅（如果reqeust通道不出现
 * 阻塞，那么其他通道也不会出现阻塞，因为request通道是所有的工作的源！）
 */
import (
	"errors"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/log"
	"github.com/hq-cml/spider-go/helper/util"
	"github.com/hq-cml/spider-go/logic/analyzer"
	"github.com/hq-cml/spider-go/logic/downloader"
	"github.com/hq-cml/spider-go/logic/processchain"
	"github.com/hq-cml/spider-go/middleware/channelmanager"
	"github.com/hq-cml/spider-go/middleware/requestcache"
	"github.com/hq-cml/spider-go/middleware/stopsign"
	"github.com/hq-cml/spider-go/middleware/pool"
	"net/http"
	"sync/atomic"
	"time"
	"strings"
)

//New
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

//统一Start的参数校验，对于入参进行逐个的校验
func (schdl *Scheduler) startParamCheck(
	context basic.Context,
	httpClient *http.Client,
	respAnalyzers []basic.AnalyzeResponseFunc,
	entryProcessors []basic.ProcessEntryFunc,
	firstHttpReq *http.Request) (error) {

	if context.Conf.GrabDepth <= 0 {
		return errors.New("GrabDepth can not be 0!")
	}

	if context.Conf.RequestChanCapcity <= 0 ||
		context.Conf.ResponseChanCapcity <= 0 ||
		context.Conf.EntryChanCapcity <= 0 ||
		context.Conf.ErrorChanCapcity <= 0 {
		return errors.New("Channel length can not be 0!")
	}

	if httpClient == nil {
		return errors.New("The httpClient can not be nil!")
	}

	if context.Conf.DownloaderPoolSize <= 0 ||
		context.Conf.AnalyzerPoolSize <= 0 {
		return errors.New("Pool size can not be 0!")
	}

	if entryProcessors == nil {
		return errors.New("The entry processor list is invalid!")
	}
	for i, ip := range entryProcessors {
		if ip == nil {
			return errors.New(fmt.Sprintf("The %dth entry processor is invalid!", i))
		}
	}

	if firstHttpReq == nil {
		return errors.New("The first HTTP request is invalid!")
	}

	return nil
}

//scheduler初始化
func (schdl *Scheduler) schedulerInit(
	context basic.Context,
	httpClient *http.Client,
	respAnalyzers []basic.AnalyzeResponseFunc,
	entryProcessors []basic.ProcessEntryFunc,
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
	if atomic.LoadUint32(&schdl.running) == RUNNING_STATUS_RUNNING {
		err = errors.New("The scheduler has been started!\n") //已经开启，则退出，单例
		return
	}
	atomic.StoreUint32(&schdl.running, RUNNING_STATUS_RUNNING)

	//GrabDepth赋值
	schdl.grabDepth = context.Conf.GrabDepth

	//middleware生成；通道管理器
	schdl.channelManager = channelmanager.NewChannelManager()
	schdl.channelManager.RegisterOneChannel("request", basic.NewRequestChannel(context.Conf.RequestChanCapcity))
	schdl.channelManager.RegisterOneChannel("response", basic.NewResponseChannel(context.Conf.ResponseChanCapcity))
	schdl.channelManager.RegisterOneChannel("entry", basic.NewEntryChannel(context.Conf.EntryChanCapcity))
	schdl.channelManager.RegisterOneChannel("error", basic.NewErrorChannel(context.Conf.ErrorChanCapcity))

	//middleware生成；池管理器
	schdl.poolManager = pool.NewPoolManager()
	if dp, err := downloader.NewDownloaderPool(context.Conf.DownloaderPoolSize,
		func() pool.EntityIntfs {
			return downloader.NewDownloader(httpClient)
		},
	); err != nil {
		err = errors.New(fmt.Sprintf("Occur error when gen downloader pool: %s\n", err))
		return err
	} else {
		//注册进入池管理器
		schdl.poolManager.RegisterOnePool("downloader", dp)
	}
	if ap, err := analyzer.NewAnalyzerPool(context.Conf.AnalyzerPoolSize, analyzer.NewAnalyzer); err != nil {
		err = errors.New(fmt.Sprintf("Occur error when gen downloader pool: %s\n", err))
		return err
	} else {
		//注册进入池管理器
		schdl.poolManager.RegisterOnePool("analyzer", ap)
	}

	//middleware生成；stopSign
	if schdl.stopSign == nil {
		schdl.stopSign = stopsign.NewStopSign()
	} else {
		schdl.stopSign.Reset()
	}

	//middleware生成；requestCache
	schdl.requestCache = requestcache.NewRequestCache()

	//processChain生成
	schdl.processChain = processchain.NewProcessChain(entryProcessors)

	//初始化已请求的URL的字典
	schdl.urlMap = make(map[string]bool)

	//主域名初始化
	if schdl.primaryDomain, err = util.GetPrimaryDomain(firstHttpReq.Host); err != nil {
		return err
	}

	return nil
}

/*
 * 开始调度，一个独立的goroutine负责：
 * 一个无限Loop，适当的搬运请求缓存中的请求到请求通道
 * 每一轮都会先计算出request通道的剩余容量，然后从缓冲中取出相同的数量的请求放入通道
 * 这样就可以完全防止request通道的阻塞，从而保证所有通道的顺畅
 */
func (schdl *Scheduler)beginToSchedule(interval time.Duration) {
	go func() {
		for {
			if schdl.stopSign.Signed() {
				schdl.stopSign.Deal(SCHEDULER_CODE)
				return
			}

			//请求通道的空闲数量（请求通道的容量 - 长度）
			remainder := schdl.getReqestChan().Cap() - schdl.getReqestChan().Len()
			var temp *basic.Request
			for remainder > 0 {
				temp = schdl.requestCache.Get()
				if temp == nil {
					break
				}

				if schdl.stopSign.Signed() {
					schdl.stopSign.Deal(SCHEDULER_CODE)
					return
				}

				schdl.getReqestChan().Put(*temp)
				remainder--
			}

			time.Sleep(interval)
		}
	}()
}

/*
 * 开启调度器。调用该方法会使调度器创建和初始化各个组件。在此之后，调度器会激活爬取流程的执行。
 * 参数httpClient是客户端句柄。
 * 参数respAnalyzers是用户定制的分析器链
 * 参数entryProcessors是用户定制的处理器链
 * 参数firstHttpReq即代表首次请求。调度器会以此为起始点开始执行爬取流程。
 */
func (schdl *Scheduler)Start(
	context basic.Context,
	httpClient *http.Client,
	respAnalyzers []basic.AnalyzeResponseFunc,
	entryProcessors []basic.ProcessEntryFunc,
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

	//统一的参数校验
	if err := schdl.startParamCheck(context, httpClient, respAnalyzers, entryProcessors, firstHttpReq); err != nil {
		return err
	}

	//初始化sheduler
	if err := schdl.schedulerInit(context, httpClient, respAnalyzers, entryProcessors, firstHttpReq); err != nil {
		return err
	}

	//下载器激活
	schdl.activateDownloaders()

	//分析器激活
	schdl.activateAnalyzers(respAnalyzers)

	//处理链激活
	schdl.activateProcessChain()

	//Error处理器激活
	schdl.activateProcessError()

	//开始调度
	schdl.beginToSchedule(10 * time.Millisecond)

	//生成第一个请求，放入请求缓冲，调度器会自动进行后续的调度。。。
	//一切的开始。。。。
	firstReq := basic.NewRequest(firstHttpReq, 0) //深度0
	schdl.requestCache.Put(firstReq)

	return nil
}

//Stop方法，该方法会停止调度器的运行。所有处理模块执行的流程都会被中止
func (schdl *Scheduler)Stop() bool {
	if atomic.LoadUint32(&schdl.running) != RUNNING_STATUS_RUNNING {
		return false
	}
	schdl.stopSign.Sign() 			//发出停止信号
	schdl.channelManager.Close()    //所有中间件关闭
	schdl.requestCache.Close()
	schdl.poolManager.Close()
	atomic.StoreUint32(&schdl.running, RUNNING_STATUS_STOP)
	return true
}

//判断调度器是否正在运行。
func (schdl *Scheduler)IsRunning() bool {
	return atomic.LoadUint32(&schdl.running) == RUNNING_STATUS_RUNNING
}

//获得错误通道。调度器以及各个处理模块运行过程中出现的所有错误都会被发送到该通道。
//若该方法的结果值为nil，则说明错误通道不可用或调度器已被停止。
func (schdl *Scheduler) ErrorChan() basic.SpiderChannelIntfs {
	//TODO 曾经出过panic地址为空的段错误
	if schdl.channelManager.Status() != channelmanager.CHANNEL_MANAGER_STATUS_INITIALIZED ||
	   schdl.poolManager.Status() != pool.POOL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	return schdl.getErrorChan()
}

//判断所有处理模块是否都处于空闲状态。
func (schdl *Scheduler) IsIdle() bool {
	idleDlPool := schdl.getDownloaderPool().Used() == 0
	idleAnalyzerPool := schdl.getAnalyzerPool().Used() == 0
	idleEntryPipeline := schdl.processChain.ProcessingNumber() == 0
	if idleDlPool && idleAnalyzerPool && idleEntryPipeline {
		return true
	}
	return false
}

//Summary方法
func (sched *Scheduler) Summary(prefix string) *SchedSummary {
	return NewSchedSummary(sched, prefix)
}

/*
 * 一些公共的函数
 */
// 获取通道管理器持有的请求通道。
func (schdl *Scheduler) getReqestChan() basic.SpiderChannelIntfs {
	requestChan, err := schdl.channelManager.GetOneChannel("request")
	if err != nil {
		panic(err)
	}
	return requestChan
}

// 获取通道管理器持有的响应通道。
func (schdl *Scheduler) getResponseChan() basic.SpiderChannelIntfs {
	respChan, err := schdl.channelManager.GetOneChannel("response")
	if err != nil {
		panic(err)
	}
	return respChan
}

// 获取通道管理器持有的条目通道。
func (schdl *Scheduler) getEntryChan() basic.SpiderChannelIntfs {
	entryChan, err := schdl.channelManager.GetOneChannel("entry")
	if err != nil {
		panic(err)
	}
	return entryChan
}

// 获取通道管理器持有的错误通道。
func (schdl *Scheduler) getErrorChan() basic.SpiderChannelIntfs {
	errorChan, err := schdl.channelManager.GetOneChannel("error")
	if err != nil {
		panic(err)
	}
	return errorChan
}

// 生成组件实例代号，比如为下载器，分析器等等生成一个全局唯一代号。
func generateModuleCode(moudle string, id uint64) string {
	return fmt.Sprintf("%s-%d", moudle, id)
}

// 解析组件实例代号。
func parseModuleCode(code string) (module, id string, err error) {
	t := strings.Split(code, "-")
	if len(t) == 2 {
		module = t[0]
		id = t[1]
	} else if len(t) == 1 {
		module = code
	} else {
		err = errors.New("code string error")
	}

	return
}

// 发送运行期间发生的各类错误到通道管理器中的错误通道
func (schdl *Scheduler) sendError(err error, mouduleCode string) bool {
	if err == nil {
		return false
	}
	module, _, e := parseModuleCode(mouduleCode)
	if e != nil {
		return false
	}

	var errorType basic.ErrorType
	switch module {
	case DOWNLOADER_CODE:
		errorType = basic.DOWNLOADER_ERROR
	case ANALYZER_CODE:
		errorType = basic.ANALYZER_ERROR
	case PROCESS_CHAIN_CODE:
		errorType = basic.ENTRY_PROCESSOR_ERROR
	}

	cError := basic.NewSpiderErr(errorType, err.Error())

	if schdl.stopSign.Signed() {
		schdl.stopSign.Deal(mouduleCode)
		//如果stop标记已经生效，则通道管理器可能已经关闭，此时不应该再进行通道写入
		return false
	}

	//错误的发送通道操作是放在goroutine异步执行的，原因是错误类型通道和其他几种通道
	//略有不同，错误通道的内容依赖调度器的使用方来读取，而其他几种择时调度器本身读取
	//所以此处需要防止由于调度器使用方不读取，不能因为这个问题阻塞了调度器本身
	go func() {
		schdl.getErrorChan().Put(cError)
	}()
	return true
}
