package scheduler

// 调度器摘要信息的接口类型
type SchedSummaryIntfs interface {
    String() string    //获得摘要信息的一般表示
    Detail() string    //获取摘要信息的详细表示
    Same(other SchedSummary) bool //判断是否与另一份摘要信息相同
}
