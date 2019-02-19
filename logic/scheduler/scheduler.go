package scheduler

/*
 * 调度器
 * 框架的核心组件，将所有的中间件和逻辑组件进行整合、同步、协调
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
	chanman "github.com/hq-cml/spider-go/middleware/channel"
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
func (schdl *Scheduler) checkParam (
	httpClient *http.Client,
	respAnalyzers []basic.AnalyzeResponseFunc,
	itemProcessors []basic.ProcessItemFunc,
	firstHttpReq *http.Request) (error) {

	if basic.Conf.GrabMaxDepth <= 0 {
		return errors.New("GrabMaxDepth can not be 0!")
	}

	if basic.Conf.RequestChanCapcity <= 0 ||
		basic.Conf.ResponseChanCapcity <= 0 ||
		basic.Conf.ItemChanCapcity <= 0 ||
		basic.Conf.ErrorChanCapcity <= 0 {
		return errors.New("Channel length can not be 0!")
	}

	if httpClient == nil {
		return errors.New("The httpClient can not be nil!")
	}

	if basic.Conf.DownloaderPoolSize <= 0 ||
		basic.Conf.AnalyzerPoolSize <= 0 {
		return errors.New("Pool size can not be 0!")
	}

	if itemProcessors == nil {
		return errors.New("The item processor list is invalid!")
	}
	for i, ip := range itemProcessors {
		if ip == nil {
			return errors.New(fmt.Sprintf("The %dth item processor is invalid!", i))
		}
	}

	if respAnalyzers == nil {
		return errors.New("The response Analyzer list is invalid!")
	}
	for i, ra := range respAnalyzers {
		if ra == nil {
			return errors.New(fmt.Sprintf("The %dth item analyzer is invalid!", i))
		}
	}

	if firstHttpReq == nil {
		return errors.New("The first HTTP request is invalid!")
	}

	return nil
}

//scheduler初始化
func (schdl *Scheduler) initScheduler(
	httpClient *http.Client,
	respAnalyzers []basic.AnalyzeResponseFunc,
	itemProcessors []basic.ProcessItemFunc,
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
	defer atomic.StoreUint32(&schdl.running, RUNNING_STATUS_RUNNING)

	//GrabDepth赋值
	schdl.grabMaxDepth = basic.Conf.GrabMaxDepth

	//middleware生成: 通道管理器

	//注册4个通道
	schdl.channelManager = chanman.NewChannelManager()
	schdl.channelManager.RegisterChannel(CHANNEL_FLAG_REQUEST,
		chanman.NewCommonChannel(basic.Conf.RequestChanCapcity, CHANNEL_FLAG_REQUEST))
	schdl.channelManager.RegisterChannel(CHANNEL_FLAG_RESPONSE,
		chanman.NewCommonChannel(basic.Conf.ResponseChanCapcity, CHANNEL_FLAG_RESPONSE))
	schdl.channelManager.RegisterChannel(CHANNEL_FLAG_ITEM,
		chanman.NewCommonChannel(basic.Conf.ItemChanCapcity, CHANNEL_FLAG_ITEM))
	schdl.channelManager.RegisterChannel(CHANNEL_FLAG_ERROR,
		chanman.NewCommonChannel(basic.Conf.ErrorChanCapcity, CHANNEL_FLAG_ERROR))

	//middleware生成: 池管理器
	schdl.poolManager = pool.NewPoolManager()

	//生成并注册downloader池子
	if dp, err := pool.NewCommonPool(
		basic.Conf.DownloaderPoolSize,
		func() basic.SpiderEntity {
			//这里是一个闭包, NewDownloader有一个参数client
			//所有的donwloader都公用同一个httpClient, 这符合golang的推荐用法
			return downloader.NewDownloader(httpClient)
		},
	); err != nil {
		err = errors.New(fmt.Sprintf("Occur error when gen downloader pool: %s\n", err))
		return err
	} else {
		//注册进入池管理器
		schdl.poolManager.RegisterPool(DOWNLOADER_CODE, dp)
	}

	//生成并注册analyzer池子
	if ap, err := pool.NewCommonPool(
		basic.Conf.AnalyzerPoolSize,
		func() basic.SpiderEntity {
			return analyzer.NewAnalyzer()
		},
	); err != nil {
		err = errors.New(fmt.Sprintf("Occur error when gen analyzer pool: %s\n", err))
		return err
	} else {
		//注册进入池管理器
		schdl.poolManager.RegisterPool(ANALYZER_CODE, ap)
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
	schdl.processChain = processchain.NewProcessChain(itemProcessors)

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
 * 一个无限Loop，适当的搬运请求缓存中的请求到请求通道, 以防止request通道的阻塞
 * 每一轮都会先计算出request通道的剩余容量，然后从缓冲中取出相同的数量的请求放入通道
 *
 * 整个框架最有可能阻塞的是request通道，因为无法预知分析出的页面会产出多少新的request
 * 如果request通道被打满阻塞，可能会导致整个框架的阻塞，所以利用request缓冲区来避免
 */
func (schdl *Scheduler)doSchedule(interval time.Duration) {
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
 * 参数respAnalyzers是用户定制的分析器列表
 * 参数itemProcessors是用户定制的处理器链
 * 参数firstHttpReq代表首次请求。调度器会以此为起始点开始执行爬取流程。
 */
func (schdl *Scheduler)Start(
	httpClient *http.Client,
	respAnalyzers []basic.AnalyzeResponseFunc,
	itemProcessors []basic.ProcessItemFunc,
	firstHttpReq *http.Request) (err error) {

	//异常兜底
	defer func() {
		if e := recover(); e != nil {
			msg := fmt.Sprintf("Fatal Scheduler Error:%s\n", e)
			log.Warn(msg)
			err = errors.New(msg)
			return
		}
	}()

	//统一的参数校验
	if err := schdl.checkParam(httpClient, respAnalyzers, itemProcessors, firstHttpReq); err != nil {
		return err
	}

	//初始化sheduler
	if err := schdl.initScheduler(httpClient, respAnalyzers, itemProcessors, firstHttpReq); err != nil {
		return err
	}

	//下载器激活
	schdl.activateDownloaders()

	//分析器激活
	schdl.activateAnalyzers(respAnalyzers)

	//处理链激活
	schdl.activateItemProcessor()

	//Error处理器激活
	schdl.activateErrorProcessor()

	//Summary打印器激活：定期打印summray报告
	schdl.activateRecordSummary()

	//开始调度
	schdl.doSchedule(10 * time.Millisecond)

	//生成第一个请求，放入请求缓冲，调度器会自动进行后续的调度。。。
	//一切的开始。。。。
	firstReq := basic.NewRequest(firstHttpReq, 0) //深度0
	schdl.requestCache.Put(firstReq)

	return nil
}

//Stop方法，停止调度器的运行。所有处理模块执行的流程都会被中止
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

//判断所有处理模块是否都处于空闲状态。
func (schdl *Scheduler) IsIdle() bool {
	idleDlPool := schdl.getDownloaderPool().Used() == 0
	idleAnalyzerPool := schdl.getAnalyzerPool().Used() == 0
	idleItemPipeline := schdl.processChain.ProcessingNumber() == 0
	if idleDlPool && idleAnalyzerPool && idleItemPipeline {
		return true
	}
	return false
}

/*
 * 一些公共的函数
 */
// 获取通道管理器持有的请求通道。
func (schdl *Scheduler) getReqestChan() basic.SpiderChannel {
	requestChan, err := schdl.channelManager.GetChannel(CHANNEL_FLAG_REQUEST)
	if err != nil {
		panic(err)
	}
	return requestChan
}

// 获取通道管理器持有的响应通道。
func (schdl *Scheduler) getResponseChan() basic.SpiderChannel {
	respChan, err := schdl.channelManager.GetChannel(CHANNEL_FLAG_RESPONSE)
	if err != nil {
		panic(err)
	}
	return respChan
}

// 获取通道管理器持有的条目通道。
func (schdl *Scheduler) getItemChan() basic.SpiderChannel {
	itemChan, err := schdl.channelManager.GetChannel(CHANNEL_FLAG_ITEM)
	if err != nil {
		panic(err)
	}
	return itemChan
}

// 获取通道管理器持有的错误通道。
func (schdl *Scheduler) getErrorChan() basic.SpiderChannel {
	errorChan, err := schdl.channelManager.GetChannel(CHANNEL_FLAG_ERROR)
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
		errorType = basic.PROCESSOR_ERROR
	}

	cError := basic.NewSpiderErr(errorType, err.Error())

	if schdl.stopSign.Signed() {
		schdl.stopSign.Deal(mouduleCode)
		//如果stop标记已经生效，则通道管理器可能已经关闭，此时不应该再进行通道写入
		return false
	}

	//不确定会出现多少Error的情况, 所以异步发送
	go func() {
		schdl.getErrorChan().Put(cError)
	}()
	return true
}
