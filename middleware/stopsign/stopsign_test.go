package stopsign

import (
	"testing"
	"time"
	"sync"
)

var sign *StopSign


func TestStopSign(t *testing.T) {
	sign = NewStopSign()
	var wg sync.WaitGroup
	wg.Add(10)
	for i:=0; i<10; i++ {
		go func(i int, wg *sync.WaitGroup) {
			for {
				if sign.Signed() {
					t.Log("go routine" , i, " Exit!")
					wg.Done()
					return
				} else {
					time.Sleep(1*time.Millisecond)
				}
			}
		}(i, &wg)
	}
	t.Log("Begin!")

	time.Sleep(10 * time.Second)
	sign.Sign()
	wg.Wait()
	t.Log()
	t.Log("Main exit!")
}