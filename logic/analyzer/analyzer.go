package analyzer

import (
	"errors"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/idgen"
	"github.com/hq-cml/spider-go/helper/log"
	"net/url"
)

/***********************************分析器**********************************/
/*
 * 分析器的作用是根据给定的分析规则链，分析指定网页内容，最终输出请求和条目：
 * 1. 条目item，是分析的最终产出结果，应该存下这个item
 * 2. 一个新的请求，如果这样的话，框架应该能够自动继续进行探测
 */
// *Analyzer实现pool.SpiderEntity接口
type Analyzer struct {
	id uint64 // ID
}
func (analyzer *Analyzer) Id() uint64 {
	return analyzer.id
}

//下载器专用的id生成器
var analyzerIdGenerator *idgen.IdGenerator = idgen.NewIdGenerator()

//New, 创建分析器
func NewAnalyzer() basic.SpiderEntity {
	id := analyzerIdGenerator.GetId()
	return &Analyzer{
		id: id,
	}
}

//AnalyzeResponseFunc是一个分析器的链，每个response都会被链上的每一个分析器分析
//返回值请求、条目、error的slice
func (analyzer *Analyzer) Analyze(
	respAnalyzeFuncs []basic.AnalyzeResponseFunc,
	resp basic.Response) ([]*basic.Item, []*basic.Request, []error) {
	//参数校验
	if respAnalyzeFuncs == nil {
		return nil, nil,[]error{errors.New("The response parser list is invalid!")}
	}

	//获取到实际的响应内容，并做校验
	httpResp := resp.HttpResp()
	if httpResp == nil {
		return nil, nil,[]error{errors.New("The http response is invalid!")}
	}

	//日志记录
	var reqUrl *url.URL = httpResp.Request.URL
	log.Infof("Parse the response (reqUrl=%s)... \n", reqUrl)

	respDepth := resp.Depth()

	//解析http响应，respAnalyzers，利用每一个分析函数进行分析
	itemList := []*basic.Item{}
	requestList := []*basic.Request{}
	errorList := []error{}
	for i, analyzeFunc := range respAnalyzeFuncs {
		if analyzeFunc == nil {
			errorList = append(errorList, errors.New(fmt.Sprintf("The document parser [%d] is invalid!", i)))
			continue
		}

		eList, rList, errList := analyzeFunc(httpResp, respDepth)

		if eList != nil && len(eList) > 0 {
			itemList = append(itemList, eList...)
		}

		if rList != nil && len(rList) > 0 {
			for _, req := range rList {
				newDepth := respDepth + 1
				if req.Depth() != newDepth { //TODO 从插件的实现来看,这个地方是不可能出现==的情况的...
					req = basic.NewRequest(req.HttpReq(), newDepth)
				}
				requestList = append(requestList, req)
			}
		}

		if errList != nil && len(errList) > 0 {
			errorList = append(errorList, errList...)
		}
	}

	return itemList, requestList, errorList
}
