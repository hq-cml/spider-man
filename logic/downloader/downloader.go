package downloader

import (
    "net/http"
    "github.com/hq-cml/spider-go/helper/id"
    "github.com/hq-cml/spider-go/basic"

)

/*
 * 网页下载器存在于下载器池中，每个下载器有自己的Id
 *
 */

var downloaderIdGenerator id.IdGeneratorIntfs = id.NewIdGenerator()

type PageDownloaderIntfs interface {
    Id() uint64 //获得下载器的Id
    Download(req basic.Request) (*basic.Response, error) //实际的下载行为
}

//网页下载器
type PageDownloader struct {
    id  uint64 //ID
    httpClient http.Client
}

//惯例New
func NewPageDownloader(client *http.Client) PageDownloaderIntfs {
    id := downloaderIdGenerator.GetId()

    if client == nil {
        client = &http.Client{}
    }

    return &PageDownloader{
        id: id,
        httpClient:*client,
    }
}

//*PageDownloader实现PageDownloaderIntfs接口
func (dl *PageDownloader) Id() int64 {
    return dl.id
}

func (dl *PageDownloader) Download(req basic.Request) (*basic.Response, error) {
    httpReq := req.HttpReq()
    httpResp, err := dl.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    return basic.NewResponse(httpResp, req.Depth()), nil
}