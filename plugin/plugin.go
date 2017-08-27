package plugin

import (
    "github.com/hq-cml/spider-go/logic/processchain"
    "github.com/hq-cml/spider-go/logic/analyzer"
    "net/http"
)

/*
 * Plugin定义
 * 定义SpiderIntfs接口，实现了这个接口的结构即可作为插件，进入Spider
 */
type SpiderIntfs interface{
    GenHttpClient() *http.Client
    GenEntryProcessors() []processchain.ProcessEntryFunc
    GenResponseAnalysers() []analyzer.AnalyzeResponseFunc
}
