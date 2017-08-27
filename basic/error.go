package basic

/*
 * Spider的错误类型
 */
import (
	"bytes"
	"fmt"
)

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
func NewSpiderErr(errType ErrorType, errMsg string) *SpiderError {
	return &SpiderError{
		errType: errType,
		errMsg:  errMsg,
	}
}
