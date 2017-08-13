package scheduler

/*
 * 调度器用到的的一些辅助的函数
 */
import (
    "github.com/hq-cml/spider-go/basic"
    "webcrawler/base"
    "fmt"
    "strings"
    "errors"
)

// 获取通道管理器持有的请求通道。
func (schdl *Scheduler) getReqestChan() chan basic.Request {
    requestChan, err := schdl.channelManager.ReqChan()
    if err != nil {
        panic(err)
    }
    return requestChan
}

// 获取通道管理器持有的响应通道。
func (schdl *Scheduler) getResponseChan() chan basic.Response {
    respChan, err := schdl.channelManager.RespChan()
    if err != nil {
        panic(err)
    }
    return respChan
}

// 获取通道管理器持有的条目通道。
func (schdl *Scheduler) getItemChan() chan basic.Item {
    itemChan, err := schdl.channelManager.ItemChan()
    if err != nil {
        panic(err)
    }
    return itemChan
}

// 获取通道管理器持有的错误通道。
func (schdl *Scheduler) getErrorChan() chan error {
    errorChan, err := schdl.channelManager.ErrorChan()
    if err != nil {
        panic(err)
    }
    return errorChan
}

// 生成组件实例代号。
func generateModuleCode(prefix string, id uint32) string {
    return fmt.Sprintf("%s-%d", prefix, id)
}

// 解析组件实例代号。
func parseModuleCode(code string) (module, id string, err error) {
    t := strings.Split(code, "-")
    if len(t) == 2 {
        module = t[0]
        id = t[1]
    }else if len(t) == 1{
        module = code
    }else{
        err = errors.New("code string error")
    }
    
    return
}

// 发送运行期间发生的各类错误
func (schdl *Scheduler) sendError(err error, mouduleCode string) bool {
    if err == nil {
        return false
    }
    module, _, e := parseModuleCode(mouduleCode)
    if e != nil {
        return false
    }

    var errorType basic.ErrorType
    switch module {
    case DOWNLOADER_CODE:
        errorType = base.DOWNLOADER_ERROR
    case ANALYZER_CODE:
        errorType = base.ANALYZER_ERROR
    case PROCESS_CHAIN_CODE:
        errorType = base.ITEM_PROCESSOR_ERROR
    }

    cError := basic.NewSpiderErr(errorType, err.Error())
    if schdl.stopSign.Signed() {
        schdl.stopSign.Deal(mouduleCode)
        return false
    }
    go func() {
        schdl.getErrorChan() <- cError
    }()
    return true
}