package main

import (
	"flag"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/config"
	"github.com/hq-cml/spider-go/helper/log"
	"github.com/hq-cml/spider-go/logic/scheduler"
	"github.com/hq-cml/spider-go/plugin"
	"net/http"
	"runtime"
	"time"
)

// 日志记录函数的类型。
// 参数level代表日志级别。级别设定：0：普通；1：警告；2：错误。
type Record func(level byte, content string)

//插件容器
var plugins = map[string]basic.SpiderPluginIntfs{
	"base": plugin.NewBaseSpider(),
	//....
}

//全局配置
var confPath *string = flag.String("c", "conf/spider.conf", "config file")
var firstUrl *string = flag.String("u", "http://www.sogou.com", "first url")

/*
 * 监视器实现：主要功能是对Scheduler的监视和控制：
 * 1. 在适当的时候停止自身和Scheduler
 * 2. 实时监控Scheduler及其各个组件的运行状况
 * 3. 一旦Scheduler及其各组件发生错误能够及时报告
 *
 *
// 参数intervalNs代表检查间隔时间，单位：纳秒。
// 参数maxIdleCount代表最大空闲计数。
// 参数autoStop被用来指示该方法是否在调度器空闲一段时间（即持续空闲时间，由intervalNs * maxIdleCount得出）之后自行停止调度器。
// 参数detailSummary被用来表示是否需要详细的摘要信息。
// 参数record代表日志记录函数。
// 当监控结束之后，该方法会会向作为唯一返回值的通道发送一个代表了空闲状态检查次数的数值。
*/
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//解析参数
	flag.Parse()

	conf, err := config.ParseConfig(*confPath)
	if err != nil {
		panic("parse conf err:" + err.Error())
	}

	context := basic.Context{
		Conf: conf,
	}

	//TODO 配置文件处理
	intervalNs := 10 * time.Millisecond
	//channelParams := basic.NewChannelParams(10, 10, 10, 10)
	//channelParams := basic.NewChannelParams(1, 1, 1, 1)     //TODO 配置
	//poolParams := basic.NewPoolParams(3, 3)
	//grabDepth := uint32(1)

	//TODO 创建日志

	//TODO 参数校验
	// 防止过小的参数值对爬取流程的影响
	if intervalNs < time.Millisecond {
		intervalNs = time.Millisecond
	}
	if conf.MaxIdleCount < 1000 {
		conf.MaxIdleCount = 1000
	}

	// 创建调度器
	schdl := scheduler.NewScheduler()

	// 监控停止通知器
	stopNotifier := make(chan byte, 1)

	//异步得从错误通道中接收和报告错误。
	AsyncReportError(schdl, record, stopNotifier)

	//记录摘要信息
	AsyncRecordSummary(schdl, false, record, stopNotifier)

	//检查空闲状态
	waitChan := AsyncLoopCheckStatus(schdl, intervalNs, conf.MaxIdleCount, true, record, stopNotifier)

	spider := plugins[conf.PluginKey]

	firstHttpReq, err := http.NewRequest("GET", *firstUrl, nil)
	if err != nil {
		log.Warnln(err.Error())
		return
	}

	schdl.Start(context,
		spider.GenHttpClient(),
		spider.GenResponseAnalysers(),
		spider.GenEntryProcessors(),
		firstHttpReq)

	//主协程同步等待
	cnt := WaitExit(waitChan)
	fmt.Println("Exit:", cnt)
}

// 检查状态，并在满足持续空闲时间的条件时采取必要措施。
func AsyncLoopCheckStatus(
	schdl *scheduler.Scheduler,
	intervalNs time.Duration,
	maxIdleCount int,
	autoStop bool,
	record Record,
	stopNotifier chan<- byte) <-chan uint64 {

	//检查计数通道
	checkCountChan := make(chan uint64, 1)

	var checkCount uint64
	// 已达到最大空闲计数的消息模板。
	var msgReachMaxIdleCount = "The scheduler has been idle for a period of time (about %s). \n" +
		"Now consider what stop it."
	// 停止调度器的消息模板。
	var msgStopScheduler = "Stop scheduler...%s."

	go func() {
		defer func() {
			stopNotifier <- 1
			stopNotifier <- 2
			checkCountChan <- checkCount
		}()
		// 等待调度器开启
		waitForSchedulerStart(schdl)
		// 准备
		var idleCount int
		var firstIdleTime time.Time
		for {
			// 检查调度器的空闲状态
			if schdl.Idle() {
				idleCount++
				if idleCount == 1 {
					firstIdleTime = time.Now()
				}
				if idleCount >= maxIdleCount {
					msg := fmt.Sprintf(msgReachMaxIdleCount, time.Since(firstIdleTime).String())
					record(0, msg)
					// 再次检查调度器的空闲状态，确保它已经可以被停止
					if schdl.Idle() {
						if autoStop {
							var result string
							if schdl.Stop() {
								result = "success"
							} else {
								result = "failing"
							}
							msg = fmt.Sprintf(msgStopScheduler, result)
							record(0, msg)
						}
						break
					} else {
						if idleCount > 0 {
							idleCount = 0
						}
					}
				}
			} else {
				if idleCount > 0 {
					idleCount = 0
				}
			}
			checkCount++
			time.Sleep(intervalNs)
		}
	}()

	return checkCountChan
}

// 记录摘要信息。
func AsyncRecordSummary(
	schdl *scheduler.Scheduler,
	detailSummary bool,
	record Record,
	stopNotifier <-chan byte) {

	// 摘要信息的模板。
	var summaryForMonitoring = "Monitor - Collected information[%d]:\n" +
		"  Goroutine number: %d\n" +
		"  Scheduler:\n%s" +
		"  Escaped time: %s\n"

	go func() {
		//阻塞等待调度器开启
		waitForSchedulerStart(schdl)

		// 准备
		var prevSchedSummary *scheduler.SchedSummary
		var prevNumGoroutine int
		var recordCount uint64 = 1
		startTime := time.Now()

		for {
			// 查看监控停止通知器
			select {
			case <-stopNotifier:
				return
			default:
			}
			// 获取摘要信息的各组成部分
			currNumGoroutine := runtime.NumGoroutine()
			currSchedSummary := schdl.Summary("    ")
			// 比对前后两份摘要信息的一致性。只有不一致时才会予以记录。主要为了防止日志的大量生产造成干扰
			if currNumGoroutine != prevNumGoroutine || !currSchedSummary.Same(prevSchedSummary) {
				schedSummaryStr := func() string {
					if detailSummary {
						return currSchedSummary.Detail()
					} else {
						return currSchedSummary.String()
					}
				}()
				// 记录摘要信息
				info := fmt.Sprintf(summaryForMonitoring,
					recordCount,
					currNumGoroutine,
					schedSummaryStr,
					time.Since(startTime).String(), //当前时间和startTime的时间间隔
				)
				record(0, info)
				prevNumGoroutine = currNumGoroutine
				prevSchedSummary = currSchedSummary
				recordCount++
			}
			//time.Sleep(time.Microsecond)
			time.Sleep(time.Second)
		}
	}()
}

//从错误通道中接收和报告错误。
func AsyncReportError(schdl *scheduler.Scheduler, record Record, stopNotifier <-chan byte) {

	go func() {
		//阻塞等待调度器开启
		waitForSchedulerStart(schdl)
		for {
			//非阻塞得查看监控停止通知器
			select {
			case <-stopNotifier:
				return
			default: //非阻塞
			}

			err, ok := schdl.ErrorChan().Get()
			if !ok {
				return
			}
			//如果errorChan关闭，则err可能是nil
			if err != nil {
				errMsg := fmt.Sprintf("Error (received from error channel): %s", err)
				record(2, errMsg)
			}
			//让出时间片
			time.Sleep(time.Microsecond)
		}
	}()
}

//阻塞等待调度器开启。
func waitForSchedulerStart(scheduler *scheduler.Scheduler) {
	for !scheduler.Running() {
		time.Sleep(time.Microsecond)
	}
}

//TODO 重构
func record(level byte, content string) {
	if content == "" {
		return
	}
	switch level {
	case 0:
		log.Infoln(content)
	case 1:
		log.Warnln(content)
	case 2:
		log.Infoln(content)
	}
}

func WaitExit(ch <-chan uint64) uint64 {
	return <-ch
}
