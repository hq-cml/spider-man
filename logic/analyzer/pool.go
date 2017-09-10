package analyzer

import (
	"github.com/hq-cml/spider-go/middleware/pool"
	"reflect"
)

func NewAnalyzerPool(total int, gen GenAnalyzerFunc) (pool.PoolIntfs, error) {
	etype := reflect.TypeOf(gen())

	pool, err := pool.NewPool(total, etype, gen)
	if err != nil {
		return nil, err
	}

	alpool := &AnalyzerPool{pool: pool, etype: etype}
	return alpool, nil
}

func (alpool *AnalyzerPool) Get() (pool.EntityIntfs, error) {
	entity, err := alpool.pool.Get()
	if err != nil {
		return nil, err
	}
	//analyzer, ok := entity.(AnalyzerIntfs)
	//if !ok {
	//	errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", alpool.etype)
	//	panic(errors.New(errMsg))
	//}
	return entity, nil
}

func (alpool *AnalyzerPool) Put(analyzer pool.EntityIntfs) error {
	return alpool.pool.Put(analyzer)
}

func (alpool *AnalyzerPool) Total() int {
	return alpool.pool.Total()
}
func (alpool *AnalyzerPool) Used() int {
	return alpool.pool.Used()
}
func (dlpool *AnalyzerPool) Close() {
	dlpool.pool.Close()
}