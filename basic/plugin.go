package basic
/*
 * Plugin定义
 * 定义SpiderPluginIntfs接口，实现了这个接口的结构即可作为插件，嵌入Spider
 */

import (
    "net/http"
)

// 被用来处理entry的函数的类型
type ProcessEntryFunc func(entry Entry) (result Entry, err error)

//被用于解析Http响应的函数的类型，这个函数类型的变量将作为参数传入Analyze，这么做
//主要是为了框架的通用性，分析规则及产出规则均可以交由用户进行自定制
//返回值是一个slice，每个成员是DataIntfs的实现，因为他们可能是上述两种情况
type AnalyzeResponseFunc func(httpResp *http.Response, respDepth int) ([]DataIntfs, []error)

/*
 * SpiderPluginIntfs接口定义
 */
type SpiderPluginIntfs interface{
    GenHttpClient() *http.Client
    GenEntryProcessors() []ProcessEntryFunc
    GenResponseAnalysers() []AnalyzeResponseFunc
}
