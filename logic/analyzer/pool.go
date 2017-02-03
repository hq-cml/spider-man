package analyzer
/*
 * 分析器池
 */

//池类型接口
type AnalyzerPoolIntfs interface {
    Get() (AnalyzerIntfs, error) // 从池中获取一个分析器
    Put(analyzer AnalyzerIntfs) error // 归还一个分析器到池子中
    Total() uint32 //获得池子总容量
    Used() uint32 //获得正在被使用的分析器数量
}
