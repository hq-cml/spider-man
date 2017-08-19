package basic
/*
 * 基本数据类型的定义
 * *Request, *Response, Item都是DataIntfs的实现
 */
import (
    "net/http"
)

/********************** 数据类型接口 ************************/
type DataIntfs interface {
    Valid() bool //数据是否有效
}

/********************** Request 相关 **********************/
//请求体结构
type Request struct {
    httpReq *http.Request   //HTTP请求的指针，为了避免零值填充和实例复制，成员用指针
    depth   uint32          //请求深度，初始请求深度是0，然后逐渐递增
}

/*********************** 响应体相关 ***********************/
//响应体结构
type Response struct {
    httpResp *http.Response  //HTTP响应的指针
    depth    uint32          //深度
}

/************************ 条目相关 ************************/
//一条响应，经过分析之后的结果，得到一个条目
type Entry map[string]interface{}

/************************ 错误类型相关 ************************/
//错误类型
type ErrorType string

//错误类型常量
const (
    DOWNLOADER_ERROR     ErrorType = "Downloader Error"
    ANALYZER_ERROR       ErrorType = "Analyzer Error"
    ITEM_PROCESSOR_ERROR ErrorType = "Item Processor Error"
)

//Spider错误接口
//实现了这个接口的类型，隐含就实现了golang自带的error接口
type SpiderErrIntfs interface {
    Type()  ErrorType //获取错误类型
    Error() string    //错误详细信息
}

//错误类型，*SpiderError实现SpiderErrIntfs接口
type SpiderError struct {
    errType    ErrorType  //错误类型
    errMsg     string     //错误信息
    fullErrMsg string     //完整错误信息
}

/*************************** 参数类型相关 ****************************/
// 参数容器的接口。
type ParamsContainerIntfs interface {
    //自检参数的有效性，并在必要时返回可以说明问题的错误值。
    Check() error
    //获得参数容器的字符串表现形式。
    String() string
}

//通道参数的容器。
type ChannelParams struct {
    reqChanLen   uint   // 请求通道的长度。
    respChanLen  uint   // 响应通道的长度。
    entryChanLen  uint   // 条目通道的长度。
    errorChanLen uint   // 错误通道的长度。
    description  string // 描述。
}

//Pool基本参数的容器。
type PoolParams struct {
    downloaderPoolSize     uint32 // 网页下载器池的尺寸。
    analyzerPoolSize       uint32 // 分析器池的尺寸。
    description            string // 描述。
}