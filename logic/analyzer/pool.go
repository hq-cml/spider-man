package analyzer

import (
	"errors"
	"fmt"
	"github.com/hq-cml/spider-go/middleware/pool"
	"reflect"
)

func NewAnalyzerPool(total uint32, gen GenAnalyzerFunc) (AnalyzerPoolIntfs, error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() pool.EntityIntfs {
		return gen()
	}
	pool, err := pool.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}

	alpool := &AnalyzerPool{pool: pool, etype: etype}
	return alpool, nil
}

func (alpool *AnalyzerPool) Get() (AnalyzerIntfs, error) {
	entity, err := alpool.pool.Get()
	if err != nil {
		return nil, err
	}
	analyzer, ok := entity.(AnalyzerIntfs)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", alpool.etype)
		panic(errors.New(errMsg))
	}
	return analyzer, nil
}

func (alpool *AnalyzerPool) Put(analyzer AnalyzerIntfs) error {
	return alpool.pool.Put(analyzer)
}

func (alpool *AnalyzerPool) Total() uint32 {
	return alpool.pool.Total()
}
func (alpool *AnalyzerPool) Used() uint32 {
	return alpool.pool.Used()
}
