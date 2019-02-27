package scheduler

import (
    "fmt"
    "errors"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/log"
    "github.com/hq-cml/spider-go/logic/analyzer"
    "github.com/hq-cml/spider-go/helper/util"
    "sync/atomic"
    "strings"
)

/*
 * 激活分析器，开始分析，分析工作由独立的goroutine进行负责，无限循环，从响应通道中获取响应
 * 每一个响应再交给独立的goroutine完成分析工作，但是goroutine并不一定能够立刻开始分析工作
 * 同时能够进行分析工作的goroutine数量, 受到分析器池容量的的制约
 */
func (schdl *Scheduler) activateAnalyzers() {
    go func() {
        for { //无限循环
            response, ok := schdl.getResponseChan().Get()
            if !ok {
                //通道已关闭
                break
            }
            resp, ok := response.(basic.Response)
            if !ok {
                continue
            }
            //启动异步分析
            go schdl.analyze(resp)
        }
    }()
}

//实际分析工作
func (schdl *Scheduler) analyze(response basic.Response) {
    atomic.AddUint64(&schdl.analyzerCnt, 1)          //原子加1
    defer atomic.AddUint64(&schdl.analyzerCnt, ^uint64(0)) //原子减1

    //异常兜底
    defer func() {
        if p := recover(); p != nil {
            msg := fmt.Sprintf("Fatal Analysis Error: %s\n", p)
            log.Err(msg)
        }
    }()

    //申请分析令牌，如果申请不到，就会阻塞等待在此处~
    entity, err := schdl.getAnalyzerPool().Get()
    if err != nil {
        //msg := fmt.Sprintf("Analyzer pool error: %s", err)
        schdl.sendError(err, ANALYZER_CODE)
        return
    }
    defer func() { //注册延时归还
        err = schdl.getAnalyzerPool().Put(entity)
        if err != nil {
            //msg := fmt.Sprintf("Analyzer pool error: %s", err)
            schdl.sendError(err, ANALYZER_CODE)
        }
    }()

    ana, ok := entity.(*analyzer.Analyzer)
    if !ok {
        msg := fmt.Sprintf("Downloader pool Wrong type")
        schdl.sendError(errors.New(msg), ANALYZER_CODE)
        return
    }

    //分析
    moudleCode := generateModuleCode(ANALYZER_CODE, ana.Id())
    itemList, requestList, errs := ana.Analyze(schdl.analyzeFuncs, response)

    //将分析出的item放到item通道里
    if itemList != nil {
        for _, item := range itemList {
            schdl.sendToItemChan(*item, moudleCode)
        }
    }

    //将分析出的request放到request缓冲
    if requestList != nil {
        for _, req := range requestList {
            schdl.sendRequestToCache(*req, moudleCode)
        }
    }

    //将错误放到错误通道里
    if errs != nil {
        for _, err := range errs {
            schdl.sendError(err, moudleCode)
        }
    }
}

//发送条目到通道管理器中的条目通道
func (schdl *Scheduler) sendToItemChan(item basic.Item, moduleCode string) bool {
    if schdl.stopSign.Signed() {
        schdl.stopSign.Deal(moduleCode)
        return false
    }
    schdl.getItemChan().Put(item)
    return true
}

//获取Pool管理器持有的分析器Pool。
func (schdl *Scheduler) getAnalyzerPool() basic.SpiderPool {
    p, err := schdl.poolManager.GetPool(ANALYZER_CODE)
    if err != nil {
        panic(err)
    }
    return p
}

//把请求存放到请求缓存。
func (schdl *Scheduler) sendRequestToCache(request basic.Request, mouduleCode string) bool {

    request.HttpReq().URL.String()

    //过滤掉非法的请求
    if schdl.filterInvalidRequest(&request) == false {
        return false
    }

    //check停止信号
    if schdl.stopSign.Signed() {
        schdl.stopSign.Deal(mouduleCode)
        return false
    }

    //请求入缓存
    schdl.requestCache.Put(&request)
    log.Debug("Send the req to Cache: ", request.HttpReq().URL.String(), "  ",
        schdl.requestCache.Length(), schdl.requestCache.Capacity())

    //标记请求; 自增请求数量
    uurl := request.HttpReq().URL.String()  //消除#和/的干扰
    uurl = strings.Split(uurl, "#")[0]
    uurl = strings.TrimRight(uurl, "/")
    schdl.urlMap.Store(uurl, basic.URL_STATUS_DOWNLOADING)
    atomic.AddUint64(&schdl.urlCnt, 1)
    return true
}

//对分析出来的请求做合法性校验，不合法的会被过滤
func (schdl *Scheduler) filterInvalidRequest(request *basic.Request) bool {
    httpRequest := request.HttpReq()
    //校验请求体本身
    if httpRequest == nil {
        log.Debugln("Ignore the request! It's HTTP request is invalid!")
        return false
    }
    requestUrl := httpRequest.URL
    if requestUrl == nil {
        log.Debugln("Ignore the request! It's url is is invalid!")
        return false
    }

    //已经处理过的URL不再处理
    uurl := requestUrl.String()  //消除#和/的干扰
    uurl = strings.Split(uurl, "#")[0]
    uurl = strings.TrimRight(uurl, "/")
    if _, ok := schdl.urlMap.Load(uurl); ok {
        log.Debugf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", requestUrl)
        return false
    }

    //如果配置只能在站内爬取, 则只有主域名相同的URL才是合法的
    if !basic.Conf.CrossSite {
        if pd, _ := util.GetPrimaryDomain(httpRequest.Host); pd != schdl.primaryDomain {
            log.Debugf("Ignore the request! It's host '%s' not in primary domain '%s'. (requestUrl=%s)\n",
                httpRequest.Host, schdl.primaryDomain, requestUrl)
            return false
        }
    }

    //请求深度不能超过阈值
    if request.Depth() > schdl.grabMaxDepth {
        log.Debugf("Ignore the request! It's depth %d greater than %d. (requestUrl=%s)\n",
            request.Depth(), schdl.grabMaxDepth, requestUrl)
        return false
    }
    return true
}
