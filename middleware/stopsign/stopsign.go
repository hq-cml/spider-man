package stopsign

import (
	"fmt"
)

//New
func NewStopSign() StopSignIntfs {
	return &StopSign{
		dealCountMap: make(map[string]uint32),
	}
}

//*StopSign实现StopSignIntfs接口
func (s *StopSign) Sign() bool {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()
	if s.signed {
		return false
	}
	s.signed = true
	return true
}

func (s *StopSign) Signed() bool {
	return s.signed
}

func (s *StopSign) Reset() {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()
	s.signed = false
	//直接重置了，扔给GC去回收了
	s.dealCountMap = make(map[string]uint32)
}

func (s *StopSign) Deal(code string) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()

	if !s.signed {
		return
	}

	if _, ok := s.dealCountMap[code]; !ok {
		s.dealCountMap[code] = 1
	} else {
		s.dealCountMap[code] += 1
	}
}

func (s *StopSign) DealCount(code string) uint32 {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()
	return s.dealCountMap[code]
}

func (s *StopSign) DealTotal() uint32 {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()
	var total uint32
	for _, v := range s.dealCountMap {
		total += v
	}
	return total
}

func (s *StopSign) Summary() string {
	if s.signed {
		return fmt.Sprintf("signed: true, dealCount: %v", s.dealCountMap)
	} else {
		return "Signed: false"
	}
}
