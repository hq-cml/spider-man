package plugin

import (
	"github.com/hq-cml/spider-man/basic"
	"net/http"
	"time"
)

/*
 * *EngineSpider实现SpiderPlugin接口
 * 此插件与搜索引擎Spider-Engine打通，爬虫爬取结果=>灌入Spider-Engine
 */
type EngineSpider struct {
	userData interface{}
}

//New
func NewEngineSpider(v interface{}) basic.SpiderPlugin {
	return &EngineSpider{
		userData: v,
	}
}

//*EngineSpider实现SpiderPlugin接口
//生成HTTP客户端
func (b *EngineSpider) GenHttpClient() *http.Client {
	//客户端必须设置一个整体超时时间，否则随着时间推移，会把downloader全部卡死
	return &http.Client{
		Transport: http.DefaultTransport,
		Timeout: time.Duration(basic.Conf.RequestTimeout) * time.Second,
	}
}

//获得响应解析函数的序列
func (b *EngineSpider) GenResponseAnalysers() []basic.AnalyzeResponseFunc {
	analyzers := []basic.AnalyzeResponseFunc {
		//闭包
		func(httpResp *basic.Response) ([]*basic.Item, []*basic.Request, []error) {
			items, reqs, errs :=  parseForATag(httpResp, b.userData)
			return items, reqs, errs
		},
	}
	return analyzers
}

// 获得条目处理链的序列。
func (b *EngineSpider) GenItemProcessors() []basic.ProcessItemFunc {
	itemProcessors := []basic.ProcessItemFunc{
		processItem,
	}
	return itemProcessors
}