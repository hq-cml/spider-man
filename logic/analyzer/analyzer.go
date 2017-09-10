package analyzer

import (
	"errors"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/idgen"
	"github.com/hq-cml/spider-go/helper/log"
	"net/url"
	"github.com/hq-cml/spider-go/middleware/pool"
	"reflect"
)

/***********************************分析器**********************************/

//下载器专用的id生成器
var analyzerIdGenerator *idgen.IdGenerator = idgen.NewIdGenerator()

//New, 创建分析器
func NewAnalyzer() pool.EntityIntfs {
	id := analyzerIdGenerator.GetId()
	return &Analyzer{
		id: id,
	}
}

//*Analyzer实现pool.EntityIntfs接口
func (analyzer *Analyzer) Id() uint64 {
	return analyzer.id
}

//AnalyzeResponseFunc是一个分析器的链，每个response都会被链上的每一个分析器分析
//返回值是一个列表，其中元素可能是两种类型：请求 or 条目
func (analyzer *Analyzer) Analyze(respAnalyzers []basic.AnalyzeResponseFunc, resp basic.Response) ([]basic.DataIntfs, []error) {
	//参数校验
	if respAnalyzers == nil {
		return nil, []error{errors.New("The response parser list is invalid!")}
	}

	//获取到实际的响应内容，并做校验
	httpResp := resp.HttpResp()
	if httpResp == nil {
		return nil, []error{errors.New("The http response is invalid!")}
	}

	//日志记录
	var reqUrl *url.URL = httpResp.Request.URL
	log.Infof("Parse the response (reqUrl=%s)... \n", reqUrl)

	respDepth := resp.Depth()

	//解析http响应，respAnalyzers，利用每一个分析函数进行分析
	dataList := []basic.DataIntfs{}
	errorList := []error{}
	for i, respAnalyzer := range respAnalyzers {
		if respAnalyzer == nil {
			errorList = append(errorList, errors.New(fmt.Sprintf("The document parser [%d] is invalid!", i)))
			continue
		}

		pDataList, pErrorList := respAnalyzer(httpResp, respDepth)

		if pDataList != nil {
			for _, pData := range pDataList {
				dataList = appendDataList(dataList, pData, respDepth)
			}
		}

		if pErrorList != nil {
			errorList = append(errorList, pErrorList...)
		}

	}

	return dataList, errorList
}

//将处理完毕的值附加到列之后
func appendDataList(dataList []basic.DataIntfs, data basic.DataIntfs, respDepth int) []basic.DataIntfs {
	//检查参数有效性
	if data == nil {
		return dataList
	}

	//断言检查data的类型
	request, ok := data.(*basic.Request)
	if ok {
		//data是请求，则封装出一个新的请求
		newDepth := respDepth + 1
		if request.Depth() != newDepth {
			request = basic.NewRequest(request.HttpReq(), newDepth)
		}
		return append(dataList, request)
	} else {
		//data是条目，则
		return append(dataList, data)
	}
}


/**********************************分析器池**********************************/

func NewAnalyzerPool(total int, gen GenAnalyzerFunc) (pool.PoolIntfs, error) {
	etype := reflect.TypeOf(gen())

	pool, err := pool.NewPool(total, etype, gen)
	if err != nil {
		return nil, err
	}

	alpool := &AnalyzerPool{pool: pool, etype: etype}
	return alpool, nil
}

func (alpool *AnalyzerPool) Get() (pool.EntityIntfs, error) {
	entity, err := alpool.pool.Get()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (alpool *AnalyzerPool) Put(analyzer pool.EntityIntfs) error {
	return alpool.pool.Put(analyzer)
}

func (alpool *AnalyzerPool) Total() int {
	return alpool.pool.Total()
}
func (alpool *AnalyzerPool) Used() int {
	return alpool.pool.Used()
}
func (dlpool *AnalyzerPool) Close() {
	dlpool.pool.Close()
}