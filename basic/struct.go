package basic

/*
 * 基本数据类型的定义
 * *Request, *Response, Entry都是DataIntfs的实现
 */
import (
	"net/http"
)

/************************************** 数据类型接口 ************************************/
//TODO 重构去掉这一块，分析结果进行拆分
type DataIntfs interface {
	Valid() bool //数据是否有效
}

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
//条目：一条Response经过分析之后的结果
//因为处理链是定制的,所以这个结构会尽量灵活以保证能够存储任意的分析结果
type Entry map[string]interface{}

/************************************** 错误类型 ***************************************/
//错误类型
type ErrorType string

//错误类型常量
const (
	DOWNLOADER_ERROR      ErrorType = "Downloader Error"
	ANALYZER_ERROR        ErrorType = "Analyzer Error"
	ENTRY_PROCESSOR_ERROR ErrorType = "Entry Processor Error"
)

//错误类型
type SpiderError struct {
	errType    ErrorType //错误类型
	errMsg     string    //错误信息
	fullErrMsg string    //完整错误信息
}

/************************************** 通道相关 *************************************/
type SpiderChannelIntfs interface {
	Put(data interface{}) error
	Get() (interface{}, bool)
	Len() int
	Cap() int
	Close()
}

//请求通道
type RequestChannel struct {
	capacity int
	reqCh    chan Request
}

//响应通道
type ResponseChannel struct {
	capacity int
	respCh   chan Response
}

//结果通道
type EntryChannel struct {
	capacity int
	entryCh  chan Entry
}

//错误通道
type ErrorChannel struct {
	capacity int
	errorCh  chan SpiderError
}

/************************************** 配置相关 *************************************/
type SpiderConf struct {
	GrabDepth int

	PluginKey string

	RequestChanCapcity  int
	ResponseChanCapcity int
	EntryChanCapcity    int
	ErrorChanCapcity    int

	DownloaderPoolSize    int
	AnalyzerPoolSize    int

	MaxIdleCount int
	IntervalNs   int
}

/************************************ Context相关 **********************************/
type Context struct {
	Conf *SpiderConf
}
