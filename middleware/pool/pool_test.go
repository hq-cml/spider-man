package pool

import (
	"testing"
	"sync"
	"github.com/hq-cml/spider-man/basic"
	"github.com/hq-cml/spider-man/helper/idgen"
	"time"
	"math/rand"
	"fmt"
)

var IdGenerator *idgen.IdGenerator = idgen.NewIdGenerator()

type Temper struct {
	id         uint64 //ID
}
func (t *Temper) Id() uint64 {
	return t.id
}

func TestPool(t *testing.T) {
	np, _ := NewCommonPool(5, func() basic.SpiderEntity {
		t := &Temper{
			id: IdGenerator.GetId(),
		}
		return t
	})
	var wg sync.WaitGroup
	wg.Add(10)
	for i:=0; i<10; i++ {
		go func(i int, wg *sync.WaitGroup) {
			entity, _ := np.Get()
			defer np.Put(entity)
			fmt.Println("Go ", i, " Got the entity: ", entity.Id())
			//随机等待3-10s
			time.Sleep(time.Duration(3 + rand.Int63n(7)) * time.Second)
			//time.Sleep(5 * time.Second)
			wg.Done()
		}(i, &wg)
	}

	wg.Wait()
	t.Log("End")
}
