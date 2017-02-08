package entitypool

import (
    "reflect"
    "fmt"
    "errors"
    "sync"
    "debug/dwarf"
)

/*
 * 实体池：池操作的抽象
 */

//实体接口类型
type Entity interface {
    Id() uint32 // ID获取方法
}

//实体池的接口类型
type PoolIntfs interface {
    Get() (Entity, error) //获取
    Put(e Entity) error   //归还
    Total() uint32        //池子容量
    Used() uint32         //池子中已使用的数量
}

//实体池类型，实现PoolIntfs接口
type Pool struct {
    total     uint32        //池容量
    etype     reflect.Type  //池子中实体的类型
    genEntity func() Entity //池中实体的生成函数
    container chan Entity   //实体容器，以channel为载体
    idContainer map[uint32]bool //实体ID容器，用于辨别一个实体有效性（是否属于该池子）
    mutex sync.Mutex  //针对IDContainer的保护锁
}

//惯例New函数，创建实体池
func NewPool(
    total uint32,
    entityType reflect.Type,
    genEntity func() Entity) (PoolIntfs, error) {

    //参数校验
    if total == 0 {
        errMsg := fmt.Sprintf("NewPool failed.(total=%d)\n", total)
        return nil, errors.New(errMsg)
    }

    //初始化容器载体channel
    size := int(total)
    container := make(chan Entity, size)
    idContainer := make(map[uint32]bool)
    for i:=0; i<size; i++ {
        newEntity := genEntity()
        if entityType != reflect.TypeOf(newEntity) {
            errMsg := fmt.Sprintf("New Pool failed. genEntity() is not %s\n", entityType)
            return nil, errors.New(errMsg)
        }
        container <- newEntity
        idContainer[newEntity.Id()] = true
    }

    pool := &Pool{
        total: total,
        etype: entityType,
        genEntity: genEntity(),
        container: container,
        idContainer: idContainer,
    }

    return pool, nil
}

//*Pool实现PoolIntfs接口
func (pool *Pool) Get() (Entity, error) {
    //channel是并发安全的，无需也不能用锁保护
    entity, ok := <-pool.container
    if !ok {
        return nil, errors.New("The inner container is invalid")
    }

    //上锁保护map，这个锁不能放到上面，有造成死锁的风险，因为channel的读取本身就可能阻塞
    pool.mutex.Lock()
    defer pool.mutex.Unlock()
    pool.idContainer[entity.Id()] = false

    return entity, nil
}
















