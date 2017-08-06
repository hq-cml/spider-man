package pool

import (
    "reflect"
    "fmt"
    "errors"
)

//惯例New函数，创建实体池
func NewPool(total uint32, entityType reflect.Type, genEntity func() EntityIntfs) (PoolIntfs, error) {
    //参数校验
    if total == 0 {
        errMsg := fmt.Sprintf("NewPool failed.(total=%d)\n", total)
        return nil, errors.New(errMsg)
    }

    //初始化容器载体channel
    size := int(total)
    container := make(chan EntityIntfs, size)
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
        genEntity: genEntity,
        container: container,
        idContainer: idContainer,
    }

    return pool, nil
}

//*Pool实现PoolIntfs接口
func (pool *Pool) Get() (EntityIntfs, error) {
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

func (pool *Pool) Put(entity EntityIntfs) error {
    //入参check：entiy不能为空
    if entity == nil {
        return errors.New("The returning entity is invalid!")
    }
    //入参check：类型需要一直
    if pool.etype != reflect.TypeOf(entity) {
        return errors.New(mt.Sprintf("The type of returning entity is NOT %s!\n", pool.etype))
    }

    entityId := entity.Id()
    tmp := pool.compareAndSetIdContainer(entityId, false, true)
    if tmp == 1 {
        //获得操作权
        pool.container <- entity //归还实体
        return nil
    } else if tmp == 0 {
        //操作权被其他goroutine得到
        return errors.New(fmt.Sprintf("The entity (id=%d) is already in the pool!\n", entityId))
    } else {
        return errors.New(fmt.Sprintf("The entity (id=%d) is illegal!\n", entityId))
    }
}

//自己实现CAS乐观锁，保护IdContainer
// 结果值：
//       -1：表示键值对不存在。
//        0：表示操作失败。其他的goroutine可能已经操作过了
//        1：表示操作成功。
func (pool *Pool) compareAndSetIdContainer(entityId uint32, oldValue bool, newValue bool) int8 {
    pool.mutex.Lock()
    pool.mutex.Unlock()

    v, ok := pool.idContainer[entityId]
    if !ok {
        return -1
    }
    if v != oldValue {
        return 0 //其他的goroutine可能已经操作过了
    }
    pool.idContainer[entityId] = newValue
    return 1 //成功获得了操作权
}

func (pool *Pool) Total() uint32 {
    return pool.total
}

func (pool *Pool) Used() uint32 {
    return pool.total - uint32(len(pool.container))
}
