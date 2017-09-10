package scheduler

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

// 开启调度器。
// 调用该方法会使调度器创建和初始化各个组件。在此之后，调度器会激活爬取流程的执行。
// 参数channelParams代表通道参数的容器。
// 参数poolParams代表池基本参数的容器。
// 参数crawlDepth代表了需要被爬取的网页的最大深度值。深度大于此值的网页会被忽略。
// 参数httpClientGenerator代表的是被用来生成HTTP客户端的函数。
// 参数respParsers的值应为分析器所需的被用来解析HTTP响应的函数的序列。
// 参数entryProcessors的值应为需要被置入条目处理管道中的条目处理器的序列。
// 参数firstHttpReq即代表首次请求。调度器会以此为起始点开始执行爬取流程。
//TODO 重构
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

	//running状态设置！
	if atomic.LoadUint32(&schdl.running) == RUNNING_STATUS_RUNNING {
		err = errors.New("The scheduler has been started!\n") //已经开启，则退出，单例
		return
	}
	atomic.StoreUint32(&schdl.running, RUNNING_STATUS_RUNNING)

	//TODO 参数校验 & 赋值
	schdl.grabDepth = context.Conf.GrabDepth

	//middleware生成
	schdl.channelManager = channelmanager.NewChannelManager()
	//TODO 参数校验 &配置参数
	schdl.channelManager.RegisterOneChannel("request", basic.NewRequestChannel(context.Conf.RequestChanCapcity))
	schdl.channelManager.RegisterOneChannel("response", basic.NewResponseChannel(context.Conf.ResponseChanCapcity))
	schdl.channelManager.RegisterOneChannel("entry", basic.NewEntryChannel(context.Conf.EntryChanCapcity))
	schdl.channelManager.RegisterOneChannel("error", basic.NewErrorChannel(context.Conf.ErrorChanCapcity))

	if httpClient == nil {
		err = errors.New("The HTTP client generator list is invalid!\n") //已经开启，则退出，单例
		return
	}

	//TODO 参数校验 & 赋值
	if schdl.downloaderPool, err = downloader.NewDownloaderPool(3,
		func() pool.EntityIntfs {
			return downloader.NewDownloader(httpClient)
		},
	); err != nil {
		err = errors.New(fmt.Sprintf("Occur error when gen downloader pool: %s\n", err))
		return
	}

	if schdl.analyzerPool, err = analyzer.NewAnalyzerPool(3,
		func() analyzer.AnalyzerIntfs {
			return analyzer.NewAnalyzer()
		},
	); err != nil {
		err = errors.New(fmt.Sprintf("Occur error when gen downloader pool: %s\n", err))
		return
	}

	if entryProcessors == nil {
		return errors.New("The entry processor list is invalid!")
	}
	for i, ip := range entryProcessors {
		if ip == nil {
			return errors.New(fmt.Sprintf("The %dth entry processor is invalid!", i))
		}
	}
	schdl.processChain = processchain.NewProcessChain(entryProcessors)

	if schdl.stopSign == nil {
		schdl.stopSign = stopsign.NewStopSign()
	} else {
		schdl.stopSign.Reset()
	}

	schdl.requestCache = requestcache.NewRequestCache()
	schdl.urlMap = make(map[string]bool)

	schdl.activateDownloaders()
	schdl.activateAnalyzers(respAnalyzers)
	schdl.activateProcessChain()
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

//实现Stop方法，调用该方法会停止调度器的运行。所有处理模块执行的流程都会被中止
//调用该方法会停止调度器的运行。所有处理模块执行的流程都会被中止。
func (schdl *Scheduler) Stop() bool {
	if atomic.LoadUint32(&schdl.running) != RUNNING_STATUS_RUNNING {
		return false
	}
	schdl.stopSign.Sign()
	schdl.channelManager.Close()
	schdl.requestCache.Close()
	atomic.StoreUint32(&schdl.running, RUNNING_STATUS_STOP)
	return true
}

//实现Running方法，判断调度器是否正在运行。
func (schdl *Scheduler) Running() bool {
	return atomic.LoadUint32(&schdl.running) == RUNNING_STATUS_RUNNING
}

//实现ErrorChan方法
//若该方法的结果值为nil，则说明错误通道不可用或调度器已被停止。
// 获得错误通道。调度器以及各个处理模块运行过程中出现的所有错误都会被发送到该通道。
// 若该方法的结果值为nil，则说明错误通道不可用或调度器已被停止。
func (schdl *Scheduler) ErrorChan() basic.SpiderChannelIntfs {
	//TODO 曾经出过panic 地址为空的段错误
	if schdl.channelManager.Status() != channelmanager.CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	return schdl.getErrorChan()
}

//实现Idle方法
//判断所有处理模块是否都处于空闲状态。
func (schdl *Scheduler) Idle() bool {
	idleDlPool := schdl.downloaderPool.Used() == 0
	idleAnalyzerPool := schdl.analyzerPool.Used() == 0
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

	analyzer, err := schdl.analyzerPool.Get()
	if err != nil {
		msg := fmt.Sprintf("Analyzer pool error: %s", err)
		schdl.sendError(errors.New(msg), SCHEDULER_CODE)
		return
	}
	defer func() { //注册延时归还
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

			switch d := data.(type) {
			case *basic.Request:
				schdl.sendRequestToCache(*d, moudleCode)
			case *basic.Entry:
				schdl.sendEntry(*d, moudleCode)
			default:
				//%T打印实际类型
				msg := fmt.Sprintf("Unsported data type:%T! (value=%v)\n", d, d)
				schdl.sendError(errors.New(msg), moudleCode)
			}
		}
	}
}

/*
 * 激活下载器，开始下载，下载工作由异步的goroutine进行负责
 * 无限循环，从请求通道中获取请求，完成下载任务
 */
func (schdl *Scheduler) activateDownloaders() {
	go func() {
		//无限循环，从请求通道中获取请求
		for {
			request, ok := schdl.getReqestChan().Get()
			if !ok {
				//通道已关闭
				break
			}
			//类型断言
			req, ok := request.(basic.Request)
			if !ok {
				break
			}
			//每个请求都交给一个独立的goroutine来处理
			go schdl.download(req)
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

	entity, err := schdl.downloaderPool.Get()
	if err != nil {
		msg := fmt.Sprintf("Downloader pool error: %s", err)
		schdl.sendError(errors.New(msg), SCHEDULER_CODE)
		return
	}
	downloader, ok := entity.(*downloader.Downloader)
	if !ok {
		msg := fmt.Sprintf("Downloader pool Wrong type")
		schdl.sendError(errors.New(msg), SCHEDULER_CODE)
		return
	}
	defer func() { //注册延时归还
		//err = schdl.downloaderPool.Put(pool.EntityIntfs(downloader))
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
