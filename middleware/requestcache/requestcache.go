package requestcache

import (
    "fmt"
    "github.com/hq-cml/spider-go/basic"
)

// 创建请求缓存。
func NewRequestCache() RequestCacheIntfs {
    rc := &RequestCache{
        cache: make([]*basic.Request, 0),
    }
    return rc
}

// *RequestCache实现RequestCacheIntfs接口
func (rc *RequestCache) Put(req *basic.Request) bool {
    if req == nil {
        return false
    }
    if rc.status == REQUEST_CACHE_STATUS_COLOSED {
        return false
    }
    rc.mutex.Lock()
    defer rc.mutex.Unlock()
    rc.cache = append(rc.cache, req)
    return true
}

func (rc *RequestCache) Get() *basic.Request {
    if rc.Length() == 0 {
        return nil
    }
    if rc.status == REQUEST_CACHE_STATUS_COLOSED {
        return nil
    }
    rc.mutex.Lock()
    defer rc.mutex.Unlock()
    req := rc.cache[0]        //从头取第一个
    rc.cache = rc.cache[1:]
    return req
}

func (rc *RequestCache) Capacity() int {
    return cap(rc.cache)
}

func (rc *RequestCache) Length() int {
    return len(rc.cache)
}

func (rc *RequestCache) Close() {
    if rc.status == REQUEST_CACHE_STATUS_COLOSED {
        return
    }
    rc.status = REQUEST_CACHE_STATUS_COLOSED
}

// 摘要信息
func (rc *RequestCache) Summary() string {
    summaryTemplate := "status: %s, " + "length: %d, " + "capacity: %d"
    summary := fmt.Sprintf(summaryTemplate,
        statusMap[rc.status],
        rc.Length(),
        rc.Capacity())
    return summary
}