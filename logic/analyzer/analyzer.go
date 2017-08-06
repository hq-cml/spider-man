package analyzer

import (
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/idgen"
    "errors"
    "fmt"
    "net/url"
    "github.com/donnie4w/go-logger/logger"
)

//下载器专用的id生成器
var analyzerIdGenerator idgen.IdGeneratorIntfs = idgen.NewIdGenerator()

//New, 创建分析器
func NewAnalyzer() AnalyzerIntfs {
    id := analyzerIdGenerator.GetId()
    return &Analyzer{
        id: uint32(id),
    }
}

//*Analyzer实现AnalyzerIntfs接口
func (analyzer *Analyzer) Id() uint32 {
    return analyzer.id
}

//把响应结果一次传递给parser函数，然后将解析的结果汇总返回
func (analyzer *Analyzer) Analyze(respAnalyzers []AnalyzeResponseFunc, resp basic.Response) ([]basic.DataIntfs, []error) {
    //参数校验
    if respAnalyzers == nil {
        return nil, []error{errors.New("The response parser list is invalid!")}
    }

    //获取到实际的响应内容，并做校验
    httpResp := resp.HttpResp()
    if httpResp == nil {
        return nil, []error{errors.New("The http response is invalid!")}
    }

    //TODO 日志记录处理了哪些url
    //var reqUrl *url.URL = httpResp.Request.URL
    //logger.Infof("Parse the response (reqUrl=%s)... \n", reqUrl)

    respDepth := resp.Depth()

    //解析http响应，respAnalyzers，利用每一个分析函数进行分析
    dataList := []basic.DataIntfs{}
    errorList := []error{}
    for i,respAnalyzer := range respAnalyzers {
        if respAnalyzer == nil {
            errorList = append(errorList, errors.New(fmt.Sprintf("The document parser [%d] is invalid!", i)))
            continue
        }

        pDataList, pErrorList := respAnalyzer(httpResp, respDepth)
    }

}

//将处理完毕的值附加到列之后
func appendDataList(dataList []basic.DataIntfs, data basic.DataIntfs, respDepth uint32) []basic.DataIntfs {
    //检查参数有效性
    if data == nil {
        return dataList
    }

    //断言检查data的类型
    req, ok := data.(*basic.Request)
    if ok {
        //data是请求

    } else {
        //data是条目，则

    }
}

