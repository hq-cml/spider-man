package basic
/*
 * Spider的错误类型
 * 实现了golang自带的error接口
 */
import (
    "bytes"
    "fmt"
)

//错误类型
type ErrorType string

//错误常量
const (
    DOWNLOADER_ERROR     ErrorType = "Downloader Error"
    ANALYZER_ERROR       ErrorType = "Analyzer Error"
    ITEM_PROCESSOR_ERROR ErrorType = "Item Processor Error"
)

//Spider错误接口
//实现了这个接口的类型，隐含就实现了error接口
type SpiderErrIntfs interface {
    Type()  ErrorType //获取错误类型
    Error() string    //错误信息
}

//错误类型，实现SpiderErrIntfs接口
type SpiderError struct {
    errType    ErrorType  //错误类型
    errMsg     string     //错误信息
    fullErrMsg string     //完整错误信息
}

//实现SpiderErrIntfs接口：错误类型
func (e *SpiderError) Type() ErrorType {
    return e.errType
}

//实现SpiderErrIntfs接口：获得错误信息
func (e *SpiderError) Error() string {
    if e.fullErrMsg == "" {
        e.genFullErrMsg()
    }
    return e.fullErrMsg
}

//生成完整错误信息
func (e *SpiderError) genFullErrMsg() {
    var buffer bytes.Buffer
    buffer.WriteString("Spider Error:")
    if e.errType != "" {
        buffer.WriteString(string(e.errType))
        buffer.WriteString(": ")
    }
    buffer.WriteString(e.errMsg)
    e.fullErrMsg = fmt.Sprintf("%s\n", buffer.String())
}

//惯例New函数
func NewSpiderErr(errType ErrorType, errMsg string) SpiderErrIntfs {
    return &SpiderError{
        errType: errType,
        errMsg : errMsg,
    }
}

