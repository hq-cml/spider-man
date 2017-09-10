package scheduler

import (
	"bytes"
	"fmt"
)

// 创建调度器摘要信息。
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
			buffer.WriteString(prefix)
			buffer.WriteString(k)
			buffer.WriteByte('\n')
		}
		urlDetail = buffer.String()
	} else {
		urlDetail = "\n"
	}
	return &SchedSummary{
		grabDepth:           schdl.grabDepth,
		chanmanSummary:      schdl.channelManager.Summary(),
		reqCacheSummary:     schdl.requestCache.Summary(),
		poolmanSummary:      schdl.poolManager.Summary(),
		processChainSummary: schdl.processChain.Summary(),
		urlCount:            urlCount,
		urlDetail:           urlDetail,
		stopSignSummary:     schdl.stopSign.Summary(),
	}
}

func (ss *SchedSummary) String() string {
	return ss.getSummary(false)
}

func (ss *SchedSummary) Detail() string {
	return ss.getSummary(true)
}

func (ss *SchedSummary) Same(other *SchedSummary) bool {
	if other == nil {
		return false
	}
	otherSs, ok := interface{}(other).(*SchedSummary)
	if !ok {
		return false
	}
	if ss.running != otherSs.grabDepth ||
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

// 获取摘要信息。
func (ss *SchedSummary) getSummary(detail bool) string {
	prefix := ss.prefix
	template := prefix + "Running: %v \n" +
		prefix + "Channel args: %s \n" +
		prefix + "Grab depth: %d \n" +
		prefix + "Channels manager: %s \n" +
		prefix + "Request cache: %s\n" +
	    prefix + "Pool manager: %s \n" +
		prefix + "Entry process chain: %s\n" +
		prefix + "Urls(%d): %s" +
		prefix + "Stop sign: %s\n"
	return fmt.Sprintf(template,
		func() bool {
			return ss.running == 1
		}(),
		ss.grabDepth,
		ss.chanmanSummary,
		ss.reqCacheSummary,
		ss.poolmanSummary,
		ss.processChainSummary,
		ss.urlCount,
		func() string {
			if detail {
				return ss.urlDetail
			} else {
				return "<concealed>\n"
			}
		}(),
		ss.stopSignSummary)
}
