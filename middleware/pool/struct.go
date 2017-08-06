package pool

import (
    "reflect"
    "sync"
)

/*
 * 实体池：池操作的抽象
 */

//实体接口类型
type EntityIntfs interface {
    Id() uint32 // ID获取方法
}

//实体池的接口类型
type PoolIntfs interface {
    Get() (EntityIntfs, error) //获取
    Put(e EntityIntfs) error   //归还
    Total() uint32        //池子容量
    Used() uint32         //池子中已使用的数量
}

//实体池类型，实现PoolIntfs接口
type Pool struct {
    total     uint32             //池容量
    etype     reflect.Type       //池子中实体的类型
    genEntity func() EntityIntfs //池中实体的生成函数
    container chan EntityIntfs   //实体容器，以channel为载体
    idContainer map[uint32]bool  //实体ID容器，用于辨别一个实体有效性（是否属于该池子）
    mutex sync.Mutex             //针对IDContainer的保护锁
}









