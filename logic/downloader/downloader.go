package downloader

import (
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/idgen"
	//"github.com/hq-cml/spider-go/middleware/pool"
	"net/http"
	//"reflect"
)

/***********************************下载器**********************************/
//网页下载器，*Downloader实现SpiderEntity接口
type Downloader struct {
	id         uint64 //ID
	httpClient http.Client
}
func (dl *Downloader) Id() uint64 {
	return dl.id
}

//下载器专用的id生成器
var downloaderIdGenerator *idgen.IdGenerator = idgen.NewIdGenerator()

//New
func NewDownloader(client *http.Client) basic.SpiderEntity {
	id := downloaderIdGenerator.GetId()

	if client == nil {
		client = &http.Client{}
	}

	return &Downloader{
		id:         id,
		httpClient: *client,
	}
}

//实际下载的工作，将http的返回结果，封装到basic.Response中
func (dl *Downloader) Download(req basic.Request) (*basic.Response, error) {
	httpReq := req.HttpReq()
	httpResp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	//TODO close? 这个地方感觉稍显怪异
	return basic.NewResponse(httpResp, req.Depth()), nil
}
