package channel

import (
	"testing"
	"sync"
	"fmt"
)

type Entity struct{
	Id int
}

func TestChannel(t *testing.T) {
	c := NewCommonChannel(5, "Test")
	var wg sync.WaitGroup

	for i:=0; i<10; i++ {
		go func(i int) {
			fmt.Println("Put ", i)
			_ = c.Put(Entity{i})
		}(i,)
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		i := 0
		for {
			r, flag := c.Get()
			if !flag {
				fmt.Println("Closed!")
				break;
			}
			e, ok := r.(Entity)
			if !ok {
				break;
			}
			i ++
			if i == 10 {
				c.Close()
			}
			fmt.Println("Got entity", e.Id)
		}
		wg.Done()
	}(&wg)

	wg.Wait()

	t.Log("End")
}
