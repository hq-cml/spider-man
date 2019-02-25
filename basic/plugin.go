package basic

/*
 * Plugin定义
 * 定义SpiderPlugin接口，实现了这个接口的结构即可作为插件，嵌入Spider框架
 */

import (
	"net/http"
)

//被用于解析Http响应的函数的类型，框架的通用性，分析规则及产出规则交由用户进行自定制
//返回值是三个
//1.是Item的slice
//2.新的request的slice
//3.第二个是错误的slice
type AnalyzeResponseFunc func(httpResp *Response) ([]*Item, []*Request, []error)

// 被用来处理item的函数的类型
type ProcessItemFunc func(item Item) (result Item, err error)

/*
 * SpiderPlugin接口定义
 */
type SpiderPlugin interface {
	//生成http的client
	GenHttpClient() 		*http.Client
	//生成分析函数链
	GenResponseAnalysers()  []AnalyzeResponseFunc
	//生成Item处理函数链
	GenItemProcessors()     []ProcessItemFunc
}
