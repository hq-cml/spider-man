package downloader

import (
	"github.com/hq-cml/spider-go/middleware/pool"
	"net/http"
	"reflect"
)

//网页下载器，*Downloader实现EntityIntfs接口
type Downloader struct {
	id         uint64 //ID
	httpClient http.Client
}

//生成网页下载器的函数的类型
type GenDownloaderFunc func() pool.EntityIntfs

/*
 * 网页下载器存在于下载器池中，每个下载器有自己的Id
 * DownloaderPool封装了pool.PoolIntfs！！
 */
type DownloaderPool struct {
	pool  pool.PoolIntfs
	etype reflect.Type
}
