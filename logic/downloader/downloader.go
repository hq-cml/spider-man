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
    Get() (PageDownloaderIntfs, error) // 从池中获取一个下载器
    Put(dl PageDownloaderIntfs) error  // 归还一个下载器到池子中
    Total() uint32                     // 获得池子总容量
    Used() uint32                      // 获得正在被使用的下载器数量
}

//网页下载器池的实现
type PageDownloaderPool struct {
    pool entitypool.Pool    //结构体的嵌套
    etype reflect.Type
}

//创建网页下载器
func NewPageDownloaderPool(total uint32, gen GenPageDownloader) (PageDownloaderPoolIntfs, error) {
    etype := reflect.TypeOf(gen())
    genEntity := func() entitypool.EntityIntfs { //函数作为一个类型的变量赋值给一个变量
        return gen()
    }
    pool, err := entitypool.NewPool(total, etype, genEntity())
    if err != nil {
        return nil, err
    }

    return &PageDownloaderPool{
        pool : pool,
        etype: etype,
    }
}

//*PageDownloaderPool实现PageDownloaderPoolIntfs
func (dlpool *PageDownloaderPool) Get() (PageDownloaderIntfs, error) {
    entity, err := dlpool.pool.Get()
    if err != nil {
        return nil, err
    }
    dl, ok := entity.(PageDownloaderIntfs)
    if !ok {
        msg := fmt.Sprintf("The type of entity is not %s!\n", dlpool.etype)
        panic(errors.New(msg))
    }

    return dl, nil
}

func (dlpool *PageDownloaderPool) Put(dl PageDownloaderIntfs) error {
    return dlpool.pool.Put(dl)
}

func (dlpool *PageDownloaderPool) Total() uint32 {
    return dlpool.pool.Total()
}

func (dlpool *PageDownloaderPool) Used() uint32 {
    return dlpool.pool.Used()
}





















