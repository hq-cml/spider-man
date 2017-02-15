package downloader

import (
    "net/http"
    "github.com/hq-cml/spider-go/helper/id"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/middleware/entitypool"
    "reflect"
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

//下载器池
//生成网页下载器的函数的类型
type GenPageDownloader func() PageDownloaderIntfs

//下载器接口类型
type PageDownloaderPoolIntfs interface {
    Get() (PageDownloaderIntfs, error)
    Put(dl PageDownloaderIntfs) error
    Total() uint32
    Used() uint32
}

//网页下载器池的实现
type DownloaderPool struct {
    pool entitypool.Pool
    etype reflect.Type
}

//创建网页下载器
func NewPageDownloaderPool(total uint32, gen GenPageDownloader) (PageDownloaderPoolIntfs, error) {
    etype := reflect.TypeOf(gen())
    genEntity := func() entitypool.Entity {
        return gen()
    }
    pool, err := entitypool.NewPool(total, etype, genEntity())
    if err != nil {
        return nil, err
    }

    return &DownloaderPool{
        pool : pool,
        etype: etype,
    }
}























