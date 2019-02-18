package pool

/*
 * 一个简单池子的实现
 * 实体池：池操作的抽象
 * 实体池中的实体的类需要实现EntityIntfs接口
 */
import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"github.com/hq-cml/spider-go/basic"
)

//实体池类型，*CommonPool实现SpiderPool接口
//用一个channel和一个map配合使用实现池子的抽象功能
type CommonPool struct {
	total       int                 		//池容量
	etype       reflect.Type        		//池子中实体的类型
	genEntity   func() basic.SpiderEntity   //池中实体的生成函数
	container   chan basic.SpiderEntity     //实体的容器，以channel为载体
	idContainer map[uint64]bool     		//实体id识别器，用于辨别一个实体有效性（是否从该池子取出，true表示在池子中，false表示不在）
	mutex       sync.Mutex                  //针对IDContainer的保护锁
}

//惯例New函数，创建实体池
func NewCommonPool(total int, entityType reflect.Type, genEntity func() basic.SpiderEntity) (basic.SpiderPool, error) {
	//参数校验
	if total == 0 {
		return nil, errors.New(fmt.Sprintf("NewPool failed.(total=%d)\n", total))
	}

	//初始化容器载体channel
	container := make(chan basic.SpiderEntity, total)
	idContainer := make(map[uint64]bool)
	for i := 0; i < total; i++ {
		newEntity := genEntity()
		if entityType != reflect.TypeOf(newEntity) {
			errMsg := fmt.Sprintf("New Pool failed. genEntity() is not %s\n", entityType)
			return nil, errors.New(errMsg)
		}
		container <- newEntity 			   //实体入池
		idContainer[newEntity.Id()] = true //占用标记
	}

	pool := &CommonPool{
		total:       total,
		etype:       entityType,
		genEntity:   genEntity,
		container:   container,
		idContainer: idContainer,
	}

	return pool, nil
}

//*Pool实现PoolIntfs接口

//取出
func (pool *CommonPool) Get() (basic.SpiderEntity, error) {
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

//归还
//用一个乐观锁保护了IdContainer, 其功能简单说就是一个实体,不能被放入池子两次
func (pool *CommonPool) Put(entity basic.SpiderEntity) error {
	//入参check：entiy不能为空
	if entity == nil {
		return errors.New("The returning entity is invalid!")
	}
	//入参check：类型需要一致
	if pool.etype != reflect.TypeOf(entity) {
		return errors.New(fmt.Sprintf("The type of returning entity is NOT %s!\n", pool.etype))
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
func (pool *CommonPool) compareAndSetIdContainer(entityId uint64, oldValue bool, newValue bool) int8 {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

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

func (pool *CommonPool) Total() int {
	return pool.total
}

func (pool *CommonPool) Used() int {
	return pool.total - len(pool.container)
}

func (pool *CommonPool) Close() {
	close(pool.container)
}
