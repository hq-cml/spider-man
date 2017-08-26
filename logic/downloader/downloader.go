package downloader

import (
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/idgen"
	"net/http"
)

/********************************下载器********************************/

//下载器专用的id生成器
var downloaderIdGenerator idgen.IdGeneratorIntfs = idgen.NewIdGenerator()

//New
func NewDownloader(client *http.Client) DownloaderIntfs {
	id := downloaderIdGenerator.GetId()

	if client == nil {
		client = &http.Client{}
	}

	return &Downloader{
		id:         uint32(id),
		httpClient: *client,
	}
}

//*Downloader实现DownloaderIntfs接口
func (dl *Downloader) Id() uint32 {
	return dl.id
}

func (dl *Downloader) Download(req basic.Request) (*basic.Response, error) {
	httpReq := req.HttpReq()
	httpResp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	return basic.NewResponse(httpResp, req.Depth()), nil
}
