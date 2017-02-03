package analyzer

import (
    "net/http"
    "github.com/hq-cml/spider-go/basic"
)

/*
 * 网页分析器存在于分析器池中，每个分析器有自己的Id
 *
 */

//被用于解析Http响应的函数的类型
type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]basic.DataIntfs, []error)

// 分析器接口类型
type AnalyzerIntfs interface {
    Id() uint32 // 获得分析器自身Id
    Analyze(respParsers []ParseResponse, resp basic.Response) ([]basic.DataIntfs, []error) //根据规则分析响应并返回请求和条目
}