package entitypool

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

