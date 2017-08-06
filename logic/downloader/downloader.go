package downloader

import (
    "net/http"
    "github.com/hq-cml/spider-go/helper/idgen"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/middleware/pool"
    "reflect"
    "fmt"
    "errors"
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
        id: uint32(id),
        httpClient:*client,
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


/********************************下载器池********************************/
//New,创建网页下载器
func NewDownloaderPool(total uint32, gen GenDownloaderFunc) (DownloaderPoolIntfs, error) {
    //直接调用gen()，利用反射获取期类型
    etype := reflect.TypeOf(gen())

    //gen()的返回值是DownloaderIntfs，但是它包含了pool.EntityIntfs所有声明方法
    //所以可以认为DownloaderIntfs是pool.EntityIntfs
    genEntity := func() pool.EntityIntfs {
        return gen()
    }
    pool, err := pool.NewPool(total, etype, genEntity())
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





















