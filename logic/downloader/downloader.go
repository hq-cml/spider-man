package downloader

import (
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/idgen"
	"github.com/hq-cml/spider-go/middleware/pool"
	"net/http"
	"reflect"
)

/***********************************下载器**********************************/
//网页下载器，*Downloader实现EntityIntfs接口
type Downloader struct {
	id         uint64 //ID
	httpClient http.Client
}

//生成网页下载器的函数的类型
type GenDownloaderFunc func() pool.EntityIntfs

//下载器专用的id生成器
var downloaderIdGenerator *idgen.IdGenerator = idgen.NewIdGenerator()

//New
func NewDownloader(client *http.Client) pool.EntityIntfs {
	id := downloaderIdGenerator.GetId()

	if client == nil {
		client = &http.Client{}
	}

	return &Downloader{
		id:         id,
		httpClient: *client,
	}
}

//*Downloader实现pool.EntityIntfs接口
func (dl *Downloader) Id() uint64 {
	return dl.id
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

/**********************************下载器池**********************************/
/*
 * 网页下载器存在于下载器池中，每个下载器有自己的Id
 * DownloaderPool封装了pool.PoolIntfs！！
 */
type DownloaderPool struct {
	pool  pool.PoolIntfs
}

//New,创建网页下载器
func NewDownloaderPool(total int, gen GenDownloaderFunc) (pool.PoolIntfs, error) {
	//直接调用gen()，利用反射获取类型
	etype := reflect.TypeOf(gen())

	pool, err := pool.NewPool(total, etype, gen)
	if err != nil {
		return nil, err
	}

	dl := &DownloaderPool {
		pool:  pool,
	}

	return dl, nil
}

//*DownloaderPool实现PoolIntfs
func (dlpool *DownloaderPool) Get() (pool.EntityIntfs, error) {
	entity, err := dlpool.pool.Get()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (dlpool *DownloaderPool) Put(dl pool.EntityIntfs) error {
	return dlpool.pool.Put(dl)
}

func (dlpool *DownloaderPool) Total() int {
	return dlpool.pool.Total()
}

func (dlpool *DownloaderPool) Used() int {
	return dlpool.pool.Used()
}

func (dlpool *DownloaderPool) Close() {
	dlpool.pool.Close()
}