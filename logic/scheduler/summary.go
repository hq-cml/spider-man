package scheduler

import (
	"bytes"
	"fmt"
	"time"
	"runtime"
	"github.com/hq-cml/spider-go/helper/log"
	"github.com/hq-cml/spider-go/basic"
	"sync/atomic"
	"strconv"
)

/*
 * 创建调度器摘要信息。
 * 汇总各个模块的summary，得到整体的摘要
 */
func NewSchedSummary(schdl *Scheduler, prefix string, detail bool) *SchedSummary {
	if schdl == nil {
		return nil
	}

	var urlDetail string
	if detail && atomic.LoadUint64(&schdl.urlCnt) > 0 {
		var buffer bytes.Buffer
		buffer.WriteByte('\n')
		schdl.urlMap.Range(func(k, v interface{}) bool { //闭包
			if v.(*basic.UrlInfo).Status != basic.URL_STATUS_DONE {
				buffer.WriteString(prefix)
				buffer.WriteString(k.(string))
				buffer.WriteString("  " + convertStatus(v.(*basic.UrlInfo).Status))
				buffer.WriteByte('\n')
			}
			return true
		})
		urlDetail = buffer.String()
	} else {
		urlDetail = "\n"
	}
	prefix = "    * " + prefix
	return &SchedSummary {
		prefix:              prefix,
		running:             schdl.running,
		grabMaxDepth:        schdl.grabMaxDepth,
		chanmanSummary:      schdl.channelManager.Summary(prefix),
		reqCacheSummary:     schdl.requestCache.Summary(prefix),
		poolmanSummary:      schdl.poolManager.Summary(prefix),
		processChainSummary: schdl.processChain.Summary(prefix),
		urlCount:            atomic.LoadUint64(&schdl.urlCnt),
		urlDetail:           urlDetail,
		stopSignSummary:     schdl.stopSign.Summary(prefix),
		analyzerCnt:   		 atomic.LoadUint64(&schdl.analyzerCnt),
		downloaderCnt:   	 atomic.LoadUint64(&schdl.downloaderCnt),
	}
}

func convertStatus(status int8) string {
	switch status {
	case basic.URL_STATUS_DOWNLOADING:
		return "下载中"
	case basic.URL_STATUS_SKIP:
		return "已跳过"
	case basic.URL_STATUS_DONE:
		return "完成"
	case basic.URL_STATUS_FATAL_ERROR:
		return "出错"
	case basic.URL_STATUS_HEAD_TIMEOUT:
		return "HEAD请求超时"
	case basic.URL_STATUS_READ_TIMEOUT:
		return "读取Body超时"
	case basic.URL_STATUS_GET_TIMEOUT:
		return "GET请求超时"
	}
	return "未知！！"
}

// 获取摘要信息。
func (ss *SchedSummary) GetSummary(detail bool) string {
	//prefix := "* " + ss.prefix
	template :=
		"    *********************************************************************\n"+
	    "    *                            SPIDER SUMMARY \n" +
		"    * Running: %v " + ". GrabDepth: %d \n" +
		"    * WorkerGoroutineNum:\n%s" +
		"    * ChannelManager:\n%s" +
		"    * PoolManager:\n%s" +
		"    * RequestCache:\n%s" +
		"    * ProcessChain:\n%s" +
		"    * StopSigin:\n%s" +
		"    * Urls(%d): %s\n" +
		"    *  \n" +
		"    *********************************************************************\n "

	d := ""
	if detail {
		d = ss.urlDetail
	} else {
		d = "<concealed>"
	}

	return fmt.Sprintf(template,
		ss.running == 1,
		ss.grabMaxDepth,
		fmt.Sprintf("    *     Downloader: %d, Analyzer: %d\n", ss.downloaderCnt, ss.analyzerCnt),
		ss.chanmanSummary,
		ss.poolmanSummary,
		ss.reqCacheSummary,
		ss.processChainSummary,
		ss.stopSignSummary,
		ss.urlCount, d)
}

// 记录摘要信息。
func (schdl *Scheduler)activateRecordSummary() {

	// 摘要信息的模板。
	var summaryForMonitoring = "\n" +
	"    Monitor - Collected information[%d]:\n" +
	"    Goroutine number: (%d)." +	" Escaped time: %s\n" +
	"    Summary:\n%s"

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

			//获取摘要信息的各组成部分
			currNumGoroutine := runtime.NumGoroutine()
			currSchedSummary := NewSchedSummary(schdl, "    ", basic.Conf.SummaryDetail)
			schedSummaryStr := currSchedSummary.GetSummary(basic.Conf.SummaryDetail)

			//记录摘要信息
			content := fmt.Sprintf(summaryForMonitoring,
				recordCount,
				currNumGoroutine,
				time.Since(startTime).String(), //当前时间和startTime的时间间隔
				schedSummaryStr,
			)
			log.Infoln(content)
			recordCount++

			//等待
			d := time.Duration(basic.Conf.SummaryInterval)
			time.Sleep(d * time.Second)
		}
	}()
}

// 记录摘要信息。
func (schdl *Scheduler)GetUrlMap() string {
	var result bytes.Buffer
	var bufDownloading bytes.Buffer
	var bufDone bytes.Buffer
	var bufSkip bytes.Buffer
	var bufError bytes.Buffer
	var bufGetTimeout bytes.Buffer
	var bufHeadTimeout bytes.Buffer
	var bufReadTimeout bytes.Buffer
	var downloadCount int64
	var doneCount int64
	var skipCount int64
	var errCount int64
	var headCount int64
	var getCount int64
	var readCount int64
	schdl.urlMap.Range(func(k, v interface{}) bool { //闭包
		switch v.(*basic.UrlInfo).Status {
		case basic.URL_STATUS_DOWNLOADING:
			downloadCount ++
			bufDownloading.WriteString("    " + k.(string))
			bufDownloading.WriteByte('\n')
		case basic.URL_STATUS_DONE:
			doneCount ++
			bufDone.WriteString("    " + k.(string))
			bufDone.WriteByte('\n')
		case basic.URL_STATUS_SKIP:
			skipCount ++
			bufSkip.WriteString("    " + k.(string) + ". Msg: " + v.(*basic.UrlInfo).Msg)
			bufSkip.WriteByte('\n')
		case basic.URL_STATUS_FATAL_ERROR:
			errCount ++
			bufError.WriteString("    " + k.(string) + ". Error: "+ v.(*basic.UrlInfo).Msg)
			bufError.WriteByte('\n')
		case basic.URL_STATUS_HEAD_TIMEOUT:
			headCount ++
			bufHeadTimeout.WriteString("    " + k.(string) + ". Error: "+ v.(*basic.UrlInfo).Msg)
			bufHeadTimeout.WriteByte('\n')
		case basic.URL_STATUS_GET_TIMEOUT:
			getCount ++
			bufGetTimeout.WriteString("    " + k.(string) + ". Error: "+ v.(*basic.UrlInfo).Msg)
			bufGetTimeout.WriteByte('\n')
		case basic.URL_STATUS_READ_TIMEOUT:
			readCount ++
			bufReadTimeout.WriteString("    " + k.(string) + ". Error: "+ v.(*basic.UrlInfo).Msg)
			bufReadTimeout.WriteByte('\n')
		}

		return true
	})
	result.WriteString("出错(" + strconv.FormatInt(errCount, 10) + ")：\n" + bufError.String())
	result.WriteString("GET请求超时(" + strconv.FormatInt(getCount, 10) + ")：\n" + bufGetTimeout.String())
	result.WriteString("HEAD请求超时(" + strconv.FormatInt(headCount, 10) + ")：\n" + bufHeadTimeout.String())
	result.WriteString("READ超时(" + strconv.FormatInt(readCount, 10) + ")：\n" + bufReadTimeout.String())
	result.WriteString("跳过(" + strconv.FormatInt(skipCount, 10) + ")：\n" + bufSkip.String())
	result.WriteString("下载中(" + strconv.FormatInt(downloadCount, 10) + ")：\n" + bufDownloading.String())
	result.WriteString("完成(" + strconv.FormatInt(doneCount, 10) + ")：\n" + bufDone.String())

	return result.String()
}