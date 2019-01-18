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

//全局配置
var confPath *string = flag.String("c", "conf/spider.conf", "config file")
var firstUrl *string = flag.String("f", "http://www.sogou.com", "first url")
//var userArgu *string = flag.String("u", "http://www.sogou.com", "user argument")

/*
 * 主函数：
 * 解析配置；初始化；启动异步调度器
 * 主协程开始主循环，主要是检查状态，并在满足持续空闲时间的条件时停止Spider
 *
 */
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//解析参数
	flag.Parse()

	//配置解析
	conf, err := config.ParseConfig(*confPath)
	if err != nil {
		panic("parse conf err:" + err.Error())
	}
	basic.Conf = conf

	//插件列表, 加载所有的支持插件
	var plugins = map[string]basic.SpiderPluginIntfs {
		"base": plugin.NewBaseSpider(),
		//....
	}

	//创建日志文件并初始化日志句柄
	log.InitLog(conf.LogPath)
	log.Infof("------------Spider Begin To Run------------")

	//配置文件处理
	intervalNs := time.Duration(conf.IntervalNs) * time.Millisecond
	//防止过小的参数值对爬取流程的影响
	if intervalNs < 10 * time.Millisecond {
		intervalNs = 10 * time.Millisecond
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
	if err := schdl.Start(spider.GenHttpClient(), spider.GenResponseAnalysers(),
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

	//TODO 信号处理

	//程序结束, 生成最终报告
	summary := scheduler.NewSchedSummary(schdl, "    ")
	log.Infoln("The Spider Finish. check times:", cnt)
	log.Infoln("Final summary:\n", summary.Detail())
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
				log.Infoln(fmt.Sprintf("The scheduler has been idle for a period of time (about %s). \n" +
					"Now it will stop!", time.Since(firstIdleTime).String()))
				//再次检查调度器的空闲状态，确保它已经可以被停止
				var result string
				if schdl.Stop() {
					result = "success"
				} else {
					result = "failing"
				}
				log.Infoln(fmt.Sprintf("Stop scheduler...%s.", result))
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

