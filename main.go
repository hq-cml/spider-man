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

	//TODO 创建日志

	//配置文件处理
	intervalNs := time.Duration(conf.IntervalNs) * time.Millisecond
	//防止过小的参数值对爬取流程的影响
	if intervalNs < 10*time.Millisecond {
		intervalNs = 10*time.Millisecond
	}
	if conf.MaxIdleCount < 5 {
		conf.MaxIdleCount = 5
	}

	//插件加载
	spider, ok := plugins[conf.PluginKey]
	if !ok {
		panic("Not found plugin:" + conf.PluginKey)
	}

	//创建并启动调度器
	schdl := scheduler.NewScheduler()
	firstHttpReq, err := http.NewRequest("GET", *firstUrl, nil)
	if err != nil {
		log.Warnln(err.Error())
		return
	}
	if err := schdl.Start(context, spider.GenHttpClient(), spider.GenResponseAnalysers(),
		spider.GenEntryProcessors(), firstHttpReq); err != nil {
		panic("Scheduler Start error:" + err.Error())
	}

	// 监控停止通知器
	//stopNotifier := make(chan byte, 1)

	//异步得从错误通道中接收和报告错误。
	//AsyncReportError(schdl, record, stopNotifier)

	//记录摘要信息
	//AsyncRecordSummary(schdl, false, record, stopNotifier)

	//主协程同步等待，检查空闲状态
	cnt := loopCheckStatus(schdl, intervalNs, conf.MaxIdleCount)

	fmt.Println("The Spider Finish. check times:", cnt)
}

//检查状态，并在满足持续空闲时间的条件时采取必要措施。
func loopCheckStatus(schdl *scheduler.Scheduler, intervalNs time.Duration,	maxIdleCount int) uint64{
	var checkCount uint64

	//等待调度器开启
	for !schdl.IsRunning() {
		time.Sleep(time.Microsecond)
	}

	var idleCount int
	var firstIdleTime time.Time
	for {
		// 检查调度器的空闲状态
		if schdl.IsIdle() {
			idleCount++
			if idleCount == 1 {
				firstIdleTime = time.Now()
			}
			if idleCount >= maxIdleCount {
				msg := fmt.Sprintf("The scheduler has been idle for a period of time (about %s). \n" +
						"Now it will stop!", time.Since(firstIdleTime).String())
				log.Infoln(msg)
				//再次检查调度器的空闲状态，确保它已经可以被停止
				var result string
				if schdl.Stop() {
					result = "success"
				} else {
					result = "failing"
				}
				msg = fmt.Sprintf("Stop scheduler...%s.", result)
				log.Infoln(msg)
				break
			}
		} else {
			idleCount = 0
		}
		checkCount++
		time.Sleep(intervalNs)
	}

	return checkCount
}

