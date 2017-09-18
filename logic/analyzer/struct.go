package analyzer

import (
	"github.com/hq-cml/spider-go/middleware/pool"
	"reflect"
)

/*
 * 分析器的作用是根据给定的分析规则链，分析指定网页内容，最终输出请求和条目：
 * 1. 条目entry，是分析的最终产出结果，应该存下这个entry
 * 2. 一个新的请求，如果这样的话，框架应该能够自动继续进行探测
 */
// 分析器接口的实现类型
type Analyzer struct {
	id uint64 // ID
}

// 生成分析器的函数类型。
type GenAnalyzerFunc func() pool.EntityIntfs

//分析器池子，AnalyzerPool嵌套了一个PoolIntfs成员
//并且，*AnalyzerPool实现了接口PoolIntfs
type AnalyzerPool struct {
	pool  pool.PoolIntfs // 实体池。
	etype reflect.Type   // 池内实体的类型。
}
