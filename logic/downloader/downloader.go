package downloader

import (
    "net/http"
    "github.com/hq-cml/spider-go/helper/id"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/middleware/entitypool"
    "reflect"
    "fmt"
    "errors"
)

//下载器专用的id生成器
var downloaderIdGenerator id.IdGeneratorIntfs = id.NewIdGenerator()

//New
func NewPageDownloader(client *http.Client) DownloaderIntfs {
    id := downloaderIdGenerator.GetId()

    if client == nil {
        client = &http.Client{}
    }

    return &Downloader{
        id: id,
        httpClient:*client,
    }
}

//*Downloader实现PageDownloaderIntfs接口
func (dl *Downloader) Id() int64 {
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



//New,创建网页下载器
func NewDownloaderPool(total uint32, gen GenDownloader) (DownloaderPoolIntfs, error) {
    etype := reflect.TypeOf(gen())
    genEntity := func() entitypool.EntityIntfs { //函数作为一个类型的变量赋值给一个变量
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

//*DownloaderPool实现DownloaderPoolIntfs
func (dlpool *DownloaderPool) Get() (DownloaderIntfs, error) {
    entity, err := dlpool.pool.Get()
    if err != nil {
        return nil, err
    }
    dl, ok := entity.(DownloaderIntfs)
    if !ok {
        msg := fmt.Sprintf("The type of entity is not %s!\n", dlpool.etype)
        panic(errors.New(msg))
    }

    return dl, nil
}

func (dlpool *DownloaderPool) Put(dl DownloaderIntfs) error {
    return dlpool.pool.Put(dl)
}

func (dlpool *DownloaderPool) Total() uint32 {
    return dlpool.pool.Total()
}

func (dlpool *DownloaderPool) Used() uint32 {
    return dlpool.pool.Used()
}





















