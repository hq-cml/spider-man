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
	_ "net/http/pprof"
	"syscall"
	"os/signal"
	"os"
)

//全局配置
var confPath *string = flag.String("c", "conf/spider.conf", "config file")
var firstUrl *string = flag.String("f", "http://www.360.cn", "first url")
var pluginName *string = flag.String("p", "base", "plugin name")
var userData *string = flag.String("u", "周鸿祎", "user argument")

/*
 * 主函数：
 * 解析配置；初始化；启动异步调度器
 * 主协程开始主循环，主要是检查状态，并在满足持续空闲时间的条件时停止Spider
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

	//启动调试器
	if conf.Pprof {
		go func() {
			http.ListenAndServe("localhost:" + conf.PprofPort, nil)
		}()
	}

	//插件列表, 加载所有的支持插件
	var plugins = map[string]basic.SpiderPlugin{
		"base": plugin.NewBaseSpider(*userData),
		//....
	}

	//创建日志文件并初始化日志句柄
	log.InitLog(conf.LogPath, conf.LogLevel)
	log.Infof("------------Spider Begin To Run------------")

	//插件指定加载
	spiderPlugin, ok := plugins[conf.PluginKey]
	if !ok {
		panic("Not found plugin:" + conf.PluginKey)
	}

	//创建首个请求
	firstHttpReq, err := http.NewRequest("GET", *firstUrl, nil)
	if err != nil {
		log.Warnln(err.Error())
		return
	}

	//创建并启动调度器
	schdl := scheduler.NewScheduler()
	if err := schdl.Start (
		spiderPlugin.GenHttpClient(),
		spiderPlugin.GenResponseAnalysers(),
		spiderPlugin.GenItemProcessors(),
		firstHttpReq); err != nil {
		panic("Scheduler Start error:" + err.Error())
	}

	//主协程同步阻塞轮训，检查空闲状态或第三方信号
	intervalNs := time.Duration(conf.IntervalNs) * time.Millisecond
	if intervalNs < 10 * time.Millisecond { //防止过小的参数值对爬取流程的影响
		intervalNs = 10 * time.Millisecond
	}
	if conf.MaxIdleCount < 5 {
		conf.MaxIdleCount = 5
	}
	cnt := loopWait(schdl, intervalNs, conf.MaxIdleCount)

	//程序结束, 生成最终报告
	summary := scheduler.NewSchedSummary(schdl, "    ", true)
	log.Infoln("The Spider Finish. check times:", cnt)
	log.Infoln("Final summary:\n", summary.GetSummary(true))
}

//检查状态，并在满足条件时采取必要退出措施。
//1. 达到了持续空闲时间
//2. 接收到了结束的信号
func loopWait(schdl *scheduler.Scheduler, intervalNs time.Duration, maxIdleCount int) uint64 {
	var checkCount uint64

	//等待调度器开启
	for !schdl.IsRunning() {
		time.Sleep(time.Microsecond)
	}

	//创建监听退出chan, 这里遇到一个坑
	//通过查阅Notify源码注释看到, 如果chan长度是0, 则select不能有defalut分支
	//因为这里有default分支,则长度不能为0, , 否则会丢失signal
	c := make(chan os.Signal, 10)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) //监听指定信号

	var idleCount int
	var firstIdleTime time.Time

	QUIT:
	for {
		//检查信号, 如果收到结束信号, 则退出
		select {
		case s := <-c:
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Infoln("Recv signal:", s, ". Begin To Stop")
				result := schdl.Stop()
				log.Infoln("Stop scheduler...", result)
				break QUIT
			default:
				log.Infoln("Recv signal: ", s)
			}
		default:
			//do nothing
			//因为存在default分支, 保证程序不会阻塞在此, 但是也要求chan os.Signal长度不能为0
		}

		//检查调度器的空闲状态, 如果满足长时间空闲阈值, 则退出
		if schdl.IsIdle() {
			idleCount++
			if idleCount == 1 {
				firstIdleTime = time.Now()
			}
			if idleCount >= maxIdleCount {
				log.Infoln(fmt.Sprintf("The scheduler has been idle for a period of time (about %s). \n" +
					"Now it will stop!", time.Since(firstIdleTime).String()))
				result := schdl.Stop()
				log.Infoln("Stop scheduler...", result)
				break QUIT
			}
		} else {
			idleCount = 0
		}

		checkCount++
		time.Sleep(intervalNs)
	}

	return checkCount
}

