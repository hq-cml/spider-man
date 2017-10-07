package scheduler

import (
	"bytes"
	"fmt"
	"time"
	"runtime"
	"github.com/hq-cml/spider-go/helper/log"
	"github.com/hq-cml/spider-go/basic"
)

/*
 * 创建调度器摘要信息。
 * 汇总各个模块的summary，得到整体的摘要
 */
func NewSchedSummary(schdl *Scheduler, prefix string) *SchedSummary {
	if schdl == nil {
		return nil
	}
	urlCount := len(schdl.urlMap)
	var urlDetail string
	if urlCount > 0 {
		var buffer bytes.Buffer
		buffer.WriteByte('\n')
		for k, _ := range schdl.urlMap {
			buffer.WriteString(prefix)
			buffer.WriteString(k)
			buffer.WriteByte('\n')
		}
		urlDetail = buffer.String()
	} else {
		urlDetail = "\n"
	}
	prefix = "* " + prefix
	return &SchedSummary{
		prefix:              prefix,
		running:             schdl.running,
		grabDepth:           schdl.grabDepth,
		chanmanSummary:      schdl.channelManager.Summary(prefix),
		reqCacheSummary:     schdl.requestCache.Summary(prefix),
		poolmanSummary:      schdl.poolManager.Summary(prefix),
		processChainSummary: schdl.processChain.Summary(prefix),
		urlCount:            urlCount,
		urlDetail:           urlDetail,
		stopSignSummary:     schdl.stopSign.Summary(prefix),
	}
}

func (ss *SchedSummary) String() string {
	return ss.getSummary(false)
}

func (ss *SchedSummary) Detail() string {
	return ss.getSummary(true)
}

// 获取摘要信息。
func (ss *SchedSummary) getSummary(detail bool) string {
	//prefix := "* " + ss.prefix
	template := "*********************************************************************\n"+
	    "*                            SPIDER SUMMARY \n" +
		"*  \n" +
		"* Running: %v \n" +
		"* GrabDepth: %d \n" +
		"* StopSigin:\n%s" +
		"* ChannelManager:\n%s" +
		"* PoolManager:\n%s" +
		"* RequestCache:\n%s" +
		"* ProcessChain:\n%s" +
		"* Urls(%d): %s\n" +
		"*  \n" +
		"*********************************************************************\n "

	d := ""
	if detail {
		d = ss.urlDetail
	} else {
		d = "<concealed>"
	}

	return fmt.Sprintf(template,
		ss.running == 1,
		ss.grabDepth,
		ss.stopSignSummary,
		ss.chanmanSummary,
		ss.poolmanSummary,
		ss.reqCacheSummary,
		ss.processChainSummary,
		ss.urlCount, d)
}


func (ss *SchedSummary) Same(other *SchedSummary) bool {
	if other == nil {
		return false
	}
	otherSs, ok := interface{}(other).(*SchedSummary)
	if !ok {
		return false
	}
	if ss.running != otherSs.running ||
	ss.grabDepth != otherSs.grabDepth ||
	ss.urlCount != otherSs.urlCount ||
	ss.stopSignSummary != otherSs.stopSignSummary ||
	ss.reqCacheSummary != otherSs.reqCacheSummary ||
	ss.processChainSummary != otherSs.processChainSummary ||
	ss.poolmanSummary != otherSs.poolmanSummary ||
	ss.chanmanSummary != otherSs.chanmanSummary {
		return false
	} else {
		return true
	}
}

// 记录摘要信息。
func (schdl *Scheduler)activateRecordSummary(context basic.Context) {

	// 摘要信息的模板。
	var summaryForMonitoring = "\n    Monitor - Collected information[%d]:\n" +
	"    Goroutine number: %d\n" +
	"    Escaped time: %s\n" +
	"    Scheduler Summary:\n%s"

	go func() {
		//准备
		var recordCount uint64 = 1
		startTime := time.Now()

		for {
			//查看监控停止通知器
			if schdl.stopSign.Signed() {
				schdl.stopSign.Deal(SUMMARY_CODE)
				return
			}

			// 获取摘要信息的各组成部分
			currNumGoroutine := runtime.NumGoroutine()
			currSchedSummary := NewSchedSummary(schdl, "    ")
			schedSummaryStr := ""
			if context.Conf.SummaryDetail {
				schedSummaryStr = currSchedSummary.Detail()
			} else {
				schedSummaryStr = currSchedSummary.String()
			}

			// 记录摘要信息
			info := fmt.Sprintf(summaryForMonitoring,
				recordCount,
				currNumGoroutine,
				time.Since(startTime).String(), //当前时间和startTime的时间间隔
				schedSummaryStr,
			)
			log.Infoln(info)
			recordCount++

			//time.Sleep(time.Microsecond)
			d := time.Duration(context.Conf.SummaryInterval)
			time.Sleep(d * time.Second)
		}
	}()
}