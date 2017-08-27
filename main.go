package main

import (
	"github.com/hq-cml/spider-go/helper/log"
	"net/http"
	"time"
	baseSpider "github.com/hq-cml/spider-go/plugin/base"
	"github.com/hq-cml/spider-go/logic/scheduler"
)

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

func main() {
	// 创建调度器
	schdl := scheduler.NewScheduler()

	//开启monitor
	intervalNs := 10 * time.Millisecond
	maxIdelCount := uint(1000)
	checkCountChan := scheduler.Monitoring(
		schdl,
		intervalNs,
		maxIdelCount,
		true,
		false,
		record,
	)

	//准备启动参数
	//channelParams := basic.NewChannelParams(10, 10, 10, 10)
	//channelParams := basic.NewChannelParams(1, 1, 1, 1)     //TODO 配置
	//poolParams := basic.NewPoolParams(3, 3)
	grabDepth := uint32(1)

	spider := baseSpider.NewBaseSpider()

	startUrl := "http://www.sogou.com"
	firstHttpReq, err := http.NewRequest("GET", startUrl, nil)
	if err != nil {
		log.Warnln(err.Error())
		return
	}

	schdl.Start(grabDepth,
		spider.GenHttpClient(),
		spider.GenResponseAnalysers(),
		spider.GenEntryProcessors(),
		firstHttpReq)

	//主协程同步等待
	<-checkCountChan
}
