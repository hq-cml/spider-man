package analyzer

import (
    "net/http"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/id"
    "errors"
    "fmt"
)

/*
 * 网页分析器存在于分析器池中，每个分析器有自己的Id
 *
 */
//下载器专用的id生成器
var analyzerIdGenerator id.IdGeneratorIntfs = id.NewIdGenerator()

//被用于解析Http响应的函数的类型
//它的返回值是一个slice，每个成员是DataIntfs的实现，因为他们可能分为两种情况：
//1. 是一个分析完毕的条目item，这样的话，应该存下这个item
//2. 也有可能是一个请求，如果这样的话，框架应该能够自动继续进行探测
type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]basic.DataIntfs, []error)

// 分析器接口类型
type AnalyzerIntfs interface {
    Id() uint64 // 获得分析器自身Id
    Analyze(respParsers []ParseResponse, resp basic.Response) ([]basic.DataIntfs, []error) //根据规则分析响应并返回请求和条目
}

// 分析器接口的实现类型
type Analyzer struct {
    id uint64 // ID
}

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
func (analyzer *Analyzer) Analyze(respParsers []ParseResponse, resp basic.Response) ([]basic.DataIntfs, []error) {
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

//分析器池类型接口
type AnalyzerPoolIntfs interface {
    Get() (AnalyzerIntfs, error)      // 从池中获取一个分析器
    Put(analyzer AnalyzerIntfs) error // 归还一个分析器到池子中
    Total() uint32                    //获得池子总容量
    Used() uint32                     //获得正在被使用的分析器数量
}