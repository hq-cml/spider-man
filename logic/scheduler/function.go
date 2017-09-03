package scheduler

/*
 * 调度器用到的的一些辅助的函数
 */
import (
	"errors"
	"fmt"
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/log"
	"github.com/hq-cml/spider-go/helper/util"
	"strings"
)

// 获取通道管理器持有的请求通道。
func (schdl *Scheduler) getReqestChan() basic.SpiderChannelIntfs {
	requestChan, err := schdl.channelManager.GetOneChannel("request")
	if err != nil {
		panic(err)
	}
	return requestChan
}

// 获取通道管理器持有的响应通道。
func (schdl *Scheduler) getResponseChan() basic.SpiderChannelIntfs {
	respChan, err := schdl.channelManager.GetOneChannel("response")
	if err != nil {
		panic(err)
	}
	return respChan
}

// 获取通道管理器持有的条目通道。
func (schdl *Scheduler) getEntryChan() basic.SpiderChannelIntfs {
	entryChan, err := schdl.channelManager.GetOneChannel("entry")
	if err != nil {
		panic(err)
	}
	return entryChan
}

// 获取通道管理器持有的错误通道。
func (schdl *Scheduler) getErrorChan() basic.SpiderChannelIntfs {
	errorChan, err := schdl.channelManager.GetOneChannel("error")
	if err != nil {
		panic(err)
	}
	return errorChan
}

// 生成组件实例代号，比如为下载器，分析器等等生成一个全局唯一代号。
func generateModuleCode(moudle string, id uint64) string {
	return fmt.Sprintf("%s-%d", moudle, id)
}

// 解析组件实例代号。
func parseModuleCode(code string) (module, id string, err error) {
	t := strings.Split(code, "-")
	if len(t) == 2 {
		module = t[0]
		id = t[1]
	} else if len(t) == 1 {
		module = code
	} else {
		err = errors.New("code string error")
	}

	return
}

// 发送运行期间发生的各类错误到通道管理器中的错误通道
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
		errorType = basic.DOWNLOADER_ERROR
	case ANALYZER_CODE:
		errorType = basic.ANALYZER_ERROR
	case PROCESS_CHAIN_CODE:
		errorType = basic.ENTRY_PROCESSOR_ERROR
	}

	cError := basic.NewSpiderErr(errorType, err.Error())

	if schdl.stopSign.Signed() {
		schdl.stopSign.Deal(mouduleCode)
		//如果stop标记已经生效，则通道管理器可能已经关闭，此时不应该再进行通道写入
		return false
	}

	//错误的发送通道操作是放在goroutine异步执行的，原因是错误类型通道和其他几种通道
	//略有不同，错误通道的内容依赖调度器的使用方来读取，而其他几种择时调度器本身读取
	//所以此处需要防止由于调度器使用方不读取，不能因为这个问题阻塞了调度器本身
	go func() {
		schdl.getErrorChan().Put(cError)
	}()
	return true
}

// 发送响应到通道管理器中的错响应通道
func (schdl *Scheduler) sendResponse(resp basic.Response, mouduleCode string) bool {
	if schdl.stopSign.Signed() {
		schdl.stopSign.Deal(mouduleCode)
		//如果stop标记已经生效，则通道管理器可能已经关闭，此时不应该再进行通道写入
		return false
	}

	schdl.getResponseChan().Put(resp)
	return true
}

//对分析出来的请求做合法性校验
func (schdl *Scheduler) filterRequest(request *basic.Request) bool {
	httpRequest := request.HttpReq()
	//校验请求体本身
	if httpRequest == nil {
		log.Warnln("Ignore the request! It's HTTP request is invalid!")
		return false
	}
	requestUrl := httpRequest.URL
	if requestUrl == nil {
		log.Warnln("Ignore the request! It's url is is invalid!")
		return false
	}

	if strings.ToLower(requestUrl.Scheme) != "http" {
		log.Warnf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", requestUrl)
		return false
	}

	//已经处理过的URL不再处理
	if _, ok := schdl.urlMap[requestUrl.String()]; ok {
		log.Warnf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", requestUrl)
		return false
	}

	//只有主域名相同的URL才是合法的
	if pd, _ := util.GetPrimaryDomain(httpRequest.Host); pd != schdl.primaryDomain {
		log.Warnf("Ignore the request! It's host '%s' not in primary domain '%s'. (requestUrl=%s)\n",
			httpRequest.Host, schdl.primaryDomain, requestUrl)
		return false
	}

	//请求深度不能超过阈值
	if request.Depth() > schdl.grabDepth {
		log.Warnf("Ignore the request! It's depth %d greater than %d. (requestUrl=%s)\n",
			request.Depth(), schdl.grabDepth, requestUrl)
		return false
	}
	return true
}

// 把请求存放到请求缓存。
func (schdl *Scheduler) sendRequestToCache(request basic.Request, mouduleCode string) bool {

	//过滤掉非法的请求
	if schdl.filterRequest(&request) == false {
		return false
	}

	//check停止信号
	if schdl.stopSign.Signed() {
		schdl.stopSign.Deal(mouduleCode)
		return false
	}

	schdl.requestCache.Put(&request)
	schdl.urlMap[request.HttpReq().URL.String()] = true

	return true
}

// 发送条目到通道管理器中的条目通道
func (schdl *Scheduler) sendEntry(entry basic.Entry, moduleCode string) bool {
	if schdl.stopSign.Signed() {
		schdl.stopSign.Deal(moduleCode)
		return false
	}
	schdl.getEntryChan().Put(entry)
	return true
}
