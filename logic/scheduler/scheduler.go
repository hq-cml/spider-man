package scheduler

/*
 * 调度器
 * 框架的核心，将所有的中间件和逻辑组件进行整合、同步、协调，组装成爬虫的核心逻辑
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
 *开启调度器。调用该方法会使调度器创建和初始化各个组件。在此之后，调度器会激活爬取流程的执行。
 * 参数httpClient是客户端句柄。
 * 参数respAnalyzers是用户定制的分析器链
 * 参数entryProcessors是用户定制的处理器链
 * 参数firstHttpReq即代表首次请求。调度器会以此为起始点开始执行爬取流程。
 */
func (schdl *Scheduler) Start(
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
	schdl.activateAnalyzers(respAnalyzers)
	schdl.activateProcessChain()
	schdl.schedule(10 * time.Millisecond)

	//生成第一个请求，放入请求缓冲，调度器会自动进行后续的调度。。。
	firstReq := basic.NewRequest(firstHttpReq, 0) //深度0
	schdl.requestCache.Put(firstReq)

	return nil
}

//实现Stop方法，调用该方法会停止调度器的运行。所有处理模块执行的流程都会被中止
//调用该方法会停止调度器的运行。所有处理模块执行的流程都会被中止。
func (schdl *Scheduler) Stop() bool {
	if atomic.LoadUint32(&schdl.running) != RUNNING_STATUS_RUNNING {
		return false
	}
	schdl.stopSign.Sign()
	schdl.channelManager.Close()
	schdl.requestCache.Close()
	schdl.poolManager.Close()
	atomic.StoreUint32(&schdl.running, RUNNING_STATUS_STOP)
	return true
}

//实现Running方法，判断调度器是否正在运行。
func (schdl *Scheduler) Running() bool {
	return atomic.LoadUint32(&schdl.running) == RUNNING_STATUS_RUNNING
}

//实现ErrorChan方法
//若该方法的结果值为nil，则说明错误通道不可用或调度器已被停止。
//获得错误通道。调度器以及各个处理模块运行过程中出现的所有错误都会被发送到该通道。
//若该方法的结果值为nil，则说明错误通道不可用或调度器已被停止。
func (schdl *Scheduler) ErrorChan() basic.SpiderChannelIntfs {
	//TODO 曾经出过panic 地址为空的段错误
	if schdl.channelManager.Status() != channelmanager.CHANNEL_MANAGER_STATUS_INITIALIZED ||
	   schdl.poolManager.Status() != pool.POOL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	return schdl.getErrorChan()
}

//实现Idle方法
//判断所有处理模块是否都处于空闲状态。
func (schdl *Scheduler) Idle() bool {
	idleDlPool := schdl.getDownloaderPool().Used() == 0
	idleAnalyzerPool := schdl.getAnalyzerPool().Used() == 0
	idleEntryPipeline := schdl.processChain.ProcessingNumber() == 0
	if idleDlPool && idleAnalyzerPool && idleEntryPipeline {
		return true
	}
	return false
}

//实现Summary方法
func (sched *Scheduler) Summary(prefix string) *SchedSummary {
	return NewSchedSummary(sched, prefix)
}

// 调度。适当的搬运请求缓存中的请求到请求通道。
func (schdl *Scheduler) schedule(interval time.Duration) {
	go func() {
		for {
			if schdl.stopSign.Signed() {
				schdl.stopSign.Deal(SCHEDULER_CODE)
				return
			}

			//请求通道的容量-长度=请求通道的空闲数量
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
 * 激活处理链
 * 一个独立的goroutine，循环从Entry通道中取出Entry
 * 然后交给独立的goroutine利用process chain去处理
 */
func (schdl *Scheduler) activateProcessChain() {
	go func() {
		schdl.processChain.SetFailFast(true)
		moudleCode := PROCESS_CHAIN_CODE
		//对一个channel进行range操作，就是循环<-操作，并且在channel关闭之后能够自动结束
		for {
			entry, ok := schdl.getEntryChan().Get()
			if !ok {
				break
			}
			e, ok := entry.(basic.Entry)
			//每次从entry通道中取出一个entry，然后扔给一个独立的gorouting处理
			go func(e basic.Entry) {
				defer func() {
					if p := recover(); p != nil {
						msg := fmt.Sprintf("Fatal entry Processing Error: %s\n", p)
						log.Warn(msg)
					}
				}()

				//放入处理链，处理链上的节点自动处理，处理完毕就不必在理会了
				errs := schdl.processChain.Send(e)
				if errs != nil {
					for _, err := range errs {
						schdl.sendError(err, moudleCode)
					}
				}
			}(e)

		}
	}()
}

/*
 * 激活分析器，开始分析，分析工作由异步的goroutine进行负责
 * 无限循环，从响应通道中获取响应，完成分析工作
 */
func (schdl *Scheduler) activateAnalyzers(respAnalyzers []basic.AnalyzeResponseFunc) {
	go func() {
		for { //无限循环
			response, ok := schdl.getResponseChan().Get()
			if !ok {
				//通道已关闭
				break
			}
			resp, ok := response.(basic.Response)
			//启动异步分析
			go schdl.analyze(respAnalyzers, resp)
		}
	}()
}

//实际分析工作
func (schdl *Scheduler) analyze(respAnalyzers []basic.AnalyzeResponseFunc, response basic.Response) {
	defer func() {
		if p := recover(); p != nil {
			msg := fmt.Sprintf("Fatal Analysis Error: %s\n", p)
			log.Warn(msg)
		}
	}()

	entity, err := schdl.getAnalyzerPool().Get()
	if err != nil {
		msg := fmt.Sprintf("Analyzer pool error: %s", err)
		schdl.sendError(errors.New(msg), SCHEDULER_CODE)
		return
	}
	defer func() { //注册延时归还
		err = schdl.getAnalyzerPool().Put(entity)
		if err != nil {
			msg := fmt.Sprintf("Analyzer pool error: %s", err)
			schdl.sendError(errors.New(msg), SCHEDULER_CODE)
		}
	}()

	//断言转换
	ana, ok := entity.(*analyzer.Analyzer)
	if !ok {
		msg := fmt.Sprintf("Downloader pool Wrong type")
		schdl.sendError(errors.New(msg), SCHEDULER_CODE)
		return
	}
	moudleCode := generateModuleCode(ANALYZER_CODE, ana.Id())
	entryList, requestList, errs := ana.Analyze(respAnalyzers, response)

	if entryList != nil {
		for _, entry := range entryList {
			schdl.sendEntry(*entry, moudleCode)
		}
	}

	if requestList != nil {
		for _, req := range requestList {
			schdl.sendRequestToCache(*req, moudleCode)
		}
	}

	if errs != nil {
		for _, err := range errs {
			schdl.sendError(err, moudleCode)
		}

	}
	//
	////Analyze返回值是一个列表，其中元素可能是两种类型：请求 or 条目
	//if dataList != nil {
	//	for _, data := range dataList {
	//		if data == nil {
	//			continue
	//		}
    //
	//		switch d := data.(type) {
	//		case *basic.Request:
	//			schdl.sendRequestToCache(*d, moudleCode)
	//		case *basic.Entry:
	//			schdl.sendEntry(*d, moudleCode)
	//		default:
	//			//%T打印实际类型
	//			msg := fmt.Sprintf("Unsported data type:%T! (value=%v)\n", d, d)
	//			schdl.sendError(errors.New(msg), moudleCode)
	//		}
	//	}
	//}
}
