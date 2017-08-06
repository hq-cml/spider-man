package analyzer

import (
    "github.com/hq-cml/spider-go/middleware/pool"
    "net/http"
    "github.com/hq-cml/spider-go/basic"
    "reflect"
)

/*
 * 分析器的作用是根据给定的规则，分析指定网页内容，最终输出请求和条目：
 * 1. 条目item，是分析的最终产出结果，应该存下这个item
 * 2. 一个新的请求，如果这样的话，框架应该能够自动继续进行探测
 */

//被用于解析Http响应的函数的类型，这个函数类型的变量将作为参数传入Analyze，这么做
//主要是为了框架的通用性，分析规则及产出规则均可以交由用户进行自定制
//返回值是一个slice，每个成员是DataIntfs的实现，因为他们可能是上述两种情况
type AnalyzeResponseFunc func(httpResp *http.Response, respDepth uint32) ([]basic.DataIntfs, []error)

// 分析器接口类型
type AnalyzerIntfs interface {
    // 获得分析器自身Id
    Id() uint32
    //根据规则分析响应并返回请求和条目
    //respParsers是一个分析器的链，每个response都会被链上的每一个分析器分析
    Analyze(respParsers []AnalyzeResponseFunc, resp basic.Response) ([]basic.DataIntfs, []error)
}

// 分析器接口的实现类型
type Analyzer struct {
    id uint32 // ID
}

//分析器池类型接口
type AnalyzerPoolIntfs interface {
    Get() (AnalyzerIntfs, error)      // 从池中获取一个分析器
    Put(analyzer AnalyzerIntfs) error // 归还一个分析器到池子中
    Total() uint32                    //获得池子总容量
    Used() uint32                     //获得正在被使用的分析器数量
}

// 生成分析器的函数类型。
type GenAnalyzerFunc func() AnalyzerIntfs

//分析器池子
type AnalyzerPool struct {
    pool  pool.Pool     // 实体池。
    etype reflect.Type // 池内实体的类型。
}