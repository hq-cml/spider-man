package requestcache

/*
 * 请求缓冲
 * 用一个带锁保护的slice实现
 */

import (
	"github.com/hq-cml/spider-go/basic"
	"sync"
)

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
