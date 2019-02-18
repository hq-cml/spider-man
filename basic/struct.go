package basic

/*
 * 基本数据类型的定义
 */
import (
	"net/http"
)

/************************************** Request ***************************************/
//请求体结构
type Request struct {
	httpReq *http.Request //HTTP请求的指针，为了避免零值填充和实例复制，成员用指针
	depth   int           //请求深度，初始请求深度是0，然后逐渐递增
}

/**************************************** 响应 ****************************************/
//响应体结构
type Response struct {
	httpResp *http.Response //HTTP响应的指针
	depth    int            //深度
}

/*************************************** 条目 *****************************************/
//条目：一条Response经过分析之后的结果（golang中map是引用类型）
//因为处理链是定制的,所以这个结构会尽量灵活以保证能够存储任意的分析结果
type Item map[string]interface{}

/************************************** 错误类型 ***************************************/
//错误类型
type ErrorType string

//错误类型常量
const (
	DOWNLOADER_ERROR      ErrorType = "Downloader Error"
	ANALYZER_ERROR        ErrorType = "Analyzer Error"
	PROCESSOR_ERROR 	  ErrorType = "Processor Error"
)

//错误类型
type SpiderError struct {
	errType    ErrorType //错误类型
	errMsg     string    //错误信息
	fullErrMsg string    //完整错误信息
}

/************************************** 通道相关 *************************************/
type SpiderChannel interface {
	Put(data interface{}) error
	Get() (interface{}, bool)
	Len() int
	Cap() int
	Close()
}

/************************************** 池子相关 *************************************/
//实体接口类型, 池中的元素
type SpiderEntity interface {
	Id() uint64 // ID获取方法
}

//实体池的接口类型
type SpiderPool interface {
	Get() (SpiderEntity, error) //从池子中获取实体
	Put(e SpiderEntity) error   //归还实体到池子
	Total() int                //池子总容量
	Used() int                 //池子中已使用的数量
	Close()                    //关闭池子的载体Channel
}

/************************************** 配置相关 *************************************/
type SpiderConf struct {
	GrabMaxDepth        int    //抓取最大深度

	PluginKey           string //插件名字，根据这个值，框架会自动选择对应的插件

	RequestChanCapcity  int    //请求通道容量
	ResponseChanCapcity int    //响应通道容量
	ItemChanCapcity     int    //条目通道容量
	ErrorChanCapcity    int    //错误通道容量

	DownloaderPoolSize  int    //下载器池大小
	AnalyzerPoolSize    int    //分析器池大小

	MaxIdleCount        int    //当满足MaxIdleCount次空闲之后，程序结束
	IntervalNs          int    //检查程序结束标志的轮训时间间隔，单位：毫秒

	SummaryDetail       bool   //是否打印详细Url
	SummaryInterval     int    //打印summary的间隔，单位：秒

	LogPath             string //日志路径
	LogLevel            string //日志级别

	Pprof				bool   //是否启动pprof
	PprofPort			string //pprof端口
}

/************************************ 全局Conf变量 **********************************/
var	Conf *SpiderConf

