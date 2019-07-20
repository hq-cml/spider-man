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
	return []basic.AnalyzeResponseFunc {
		parse360NewsPage,
	}
}

// 获得条目处理链的序列。
func (b *EngineSpider) GenItemProcessors() []basic.ProcessItemFunc {
	return []basic.ProcessItemFunc{
		//闭包
		func(item basic.Item) (basic.Item, error) {
			addr, ok := b.userData.(string)
			if !ok {
				panic("Wrong type")
			}
			return processEngineItem(item, addr)
		},

	}
}

// 页面分析
// 针对360的新闻页面进行分析
// 360新闻首页：http://www.360.cn/news.html
// 常规新闻URL：http://www.360.cn/n/10758.html
func parse360NewsPage(httpResp *basic.Response) ([]*basic.Item, []*basic.Request, []error) {


	return nil, nil, nil
}

// 条目处理函数
// 发送到Spider-Engine
func processEngineItem(item basic.Item, engineAddr string) (result basic.Item, err error) {


	return nil, nil
}