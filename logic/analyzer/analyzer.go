package analyzer

import (
    "net/http"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/id"
    "errors"
    "fmt"
)

//下载器专用的id生成器
var analyzerIdGenerator id.IdGeneratorIntfs = id.NewIdGenerator()

//创建分析器
func NewAnalyzer() AnalyzerIntfs {
    return &Analyzer{
        id: analyzerIdGenerator.GetId(),
    }
}

//*Analyzer实现AnalyzerIntfs接口
func (analyzer *Analyzer) Id() uint64 {
    return analyzer.id
}

//把响应结果一次传递给parser函数，然后将解析的结果汇总返回
func (analyzer *Analyzer) Analyze(respParsers []ParseResponseFunc, resp basic.Response) ([]basic.DataIntfs, []error) {
    //参数校验
    if respParsers == nil {
        return nil, []error{errors.New("The response parser list is invalid!")}
    }

    //获取到实际的响应内容，并做校验
    httpResp := resp.HttpResp()
    if httpResp == nil {
        return nil, []error{errors.New("The http response is invalid!")}
    }

    //TODO 日志记录处理了哪些url

    respDepth := resp.Depth()

    //解析http响应，循环遍历respParsers，利用每一个分析函数进行分析
    dataList := []basic.DataIntfs{}
    errorList := []error{}
    for i,respParser := range respParsers {
        if respParser == nil {
            errorList = append(errorList, errors.New(fmt.Sprintf("The document parser [%d] is invalid!", i)))
            continue
        }

        pDataList, pErrorList := respParser(httpResp, respDepth)
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

