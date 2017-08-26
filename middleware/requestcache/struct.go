package requestcache

/*
 * 请求缓冲
 * 用一个带锁保护的slice组成
 */

import (
	"github.com/hq-cml/spider-go/basic"
	"sync"
)

// 请求缓存的接口类型。
type RequestCacheIntfs interface {
	Put(req *basic.Request) bool // 将请求放入请求缓存。
	Get() *basic.Request         // 从请求缓存获取最早被放入且仍在其中的请求。
	Capacity() int               // 获得请求缓存的容量。
	Length() int                 // 获得请求缓存的实时长度，即：其中的请求的即时数量。
	Close()                      // 关闭请求缓存。
	Summary() string             // 获取请求缓存的摘要信息。
}

// 请求缓存的实现类型。
type RequestCache struct {
	cache  []*basic.Request // 请求的存储介质。
	mutex  sync.Mutex       // 互斥锁。
	status int              // 缓存状态。0表示正在运行，1表示已关闭。
}

const (
	REQUEST_CACHE_STATUS_RUNNING = 0
	REQUEST_CACHE_STATUS_COLOSED = 1
)

//状态字典。
var statusMap = map[int]string{
	REQUEST_CACHE_STATUS_RUNNING: "running",
	REQUEST_CACHE_STATUS_COLOSED: "closed",
}
