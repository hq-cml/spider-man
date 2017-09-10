package pool
/*
 * Pool管理器
 * 框架用到两种类型的pool：分析器池 & 下载器池
 * 他们均是PoolIntfs接口的实现类型
 */
import (
    "bytes"
    "errors"
    "fmt"
    "sync"
)

//通道管理器的状态的类型。
type PoolManagerStatus uint8

const (
    POOL_MANAGER_STATUS_UNINITIALIZED PoolManagerStatus = iota     //未初始化
    POOL_MANAGER_STATUS_INITIALIZED                                //已完成初始化
    POOL_MANAGER_STATUS_CLOSED                                     //已关闭
)

//状态码与状态名称映射字典
var statusNameMap = map[PoolManagerStatus]string{
    POOL_MANAGER_STATUS_UNINITIALIZED: "uninitialized",
    POOL_MANAGER_STATUS_INITIALIZED:   "inititalized",
    POOL_MANAGER_STATUS_CLOSED:        "closed",
}

//Pool管理器实现类型
type PoolManager struct {
    pools   map[string]PoolIntfs //通道容器
    status  PoolManagerStatus    //channel管理器状态
    rwmutex sync.RWMutex         //读写锁
}

//New
func NewPoolManager() *PoolManager {
    pm := &PoolManager{
        status:  POOL_MANAGER_STATUS_UNINITIALIZED,
        pools: make(map[string]PoolIntfs),
    }
    return pm
}

//关闭通道管理器
func (pm *PoolManager) Close() bool {
    //写锁保护
    pm.rwmutex.Lock()
    defer pm.rwmutex.Unlock()

    //状态校验
    if pm.status != POOL_MANAGER_STATUS_INITIALIZED {
        return false
    }

    //逐个关闭
    for _, p := range pm.pools {
        p.Close()
    }
    pm.status = POOL_MANAGER_STATUS_CLOSED

    return true
}

//注册一个新的通道进入管理器
func (pm *PoolManager) RegisterOnePool(name string, c PoolIntfs) error {
    //写锁保护
    pm.rwmutex.Lock()
    defer pm.rwmutex.Unlock()

    if _, ok := pm.pools[name]; ok {
        return errors.New("Already Exist pool")
    }
    pm.pools[name] = c

    return nil
}

//获取request通道
func (pm *PoolManager) GetOnePool(name string) (PoolIntfs, error) {
    //读锁保护
    pm.rwmutex.RLock()
    defer pm.rwmutex.RUnlock()

    c, ok := pm.pools[name]
    if !ok {
        return nil, errors.New("Not found")
    }

    return c, nil
}

//摘要方法
func (pm *PoolManager) Summary() string {
    //读锁保护
    pm.rwmutex.RLock()
    defer pm.rwmutex.RUnlock()

    var buff bytes.Buffer
    buff.WriteString("PoolManager Status:" + statusNameMap[pm.status] + "\n")
    for k, p := range pm.pools {
        buff.WriteString(fmt.Sprint("%s: Total:%d, Used:/%d\n ", k, p.Total(), p.Used()))
    }

    return buff.String()
}

func (pm *PoolManager) Status() PoolManagerStatus {
    return pm.status
}
