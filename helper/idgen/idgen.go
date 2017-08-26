package idgen

import (
	"math"
	"sync"
)

//ID 生成器接口
type IdGeneratorIntfs interface {
	GetId() uint64 //获得一个uint64类型的ID
}

type IdGenerator struct {
	sn    uint64
	ended bool //是否已经达到最大的值
	mutex sync.Mutex
}

//惯例New
func NewIdGenerator() IdGeneratorIntfs {
	return &IdGenerator{}
}

//*IdGenerator实现IdGeneratorIntfs接口
func (gen *IdGenerator) GetId() uint64 {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	if gen.ended {
		gen.sn = 0
		gen.ended = false
		return gen.sn
	}

	id := gen.sn
	if id < math.MaxUint64 {
		gen.sn++
	} else {
		gen.ended = true
	}

	return id
}
