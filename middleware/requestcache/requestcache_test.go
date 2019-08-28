package requestcache

import (
	"testing"
	"sync"
	"github.com/hq-cml/spider-man/basic"
	"reflect"
)

func TestReqcache(t *testing.T) {
	rc := NewRequestCache()
	var wg sync.WaitGroup
	wg.Add(10)
	for i:=0; i<10; i++ {
		go func(i int, wg *sync.WaitGroup) int{
			rc.Put(basic.NewRequest(nil, i))
			wg.Done()
			return 1
		}(i, &wg)
	}

	wg.Wait()
	i:=0
	var tmp *basic.Request
	for {
		r := rc.Get()
		if reflect.TypeOf(tmp) != reflect.TypeOf(r) {
			panic("Type wrong")
		}
		t.Log("Got req", i, )
		i ++
		if rc.Length() == 0 {
			break;
		}
	}

	t.Log("End")
}
