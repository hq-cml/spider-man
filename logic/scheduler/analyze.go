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
    "net/http"
)

/*
 * 激活分析器，开始分析，分析工作由独立的goroutine进行负责，无限循环，从响应通道中获取响应
 * 每一个响应再交给独立的goroutine完成分析工作，但是goroutine并不一定能够立刻开始分析工作
 * 同时能够进行分析工作的goroutine数量, 受到分析器池容量的的制约
 *
 * 对于池子的使用，analyzer和downloader略有不同:
 * Downloader将取令牌操作，放在新建goroutine之外，原因是大概率下，Request数都非常大，如果如果放在里面会导致产生大量的等待的goroutine
 *      （因为Request Chan有Request Cache保护，所以不会出现全局阻塞，所以这么处理大面积降低了goroutine数量
 *
 * Analizer则不同，Response Chan没有特殊的保护，所以不能让其满了进而阻塞全局。并且由于通常Analyze程序是CPU型（DownLoader是IO型）
 *      很快都能够结束，所以直接将取令牌操作放在了
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
            msg := fmt.Sprintf("Fatal Analysis Error:%s. (ReqUrl=%s)\n", p, response.ReqUrl)
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
            schdl.sendRequestToCache(req, moudleCode, response.ReqUrl)
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
func (schdl *Scheduler) sendRequestToCache(request *basic.Request, mouduleCode, refUrl string) bool {

    //消除#和/的干扰, 如有必要，则重建request
    var req *basic.Request
    uurl := request.HttpReq().URL.String()
    uurl = strings.Split(uurl, "#")[0]
    uurl = strings.TrimRight(uurl, "/")
    if (uurl != request.HttpReq().URL.String()) { //
        httpReq, err := http.NewRequest(http.MethodGet, uurl, nil)
        if err != nil {
            return false
        }
        req = basic.NewRequest(httpReq, request.Depth())
    } else {
        req = request
    }

    //过滤掉非法的或者重复的请求
    if schdl.filterInvalidRequest(req) == false {
        return false
    }

    //check停止信号
    if schdl.stopSign.Signed() {
        schdl.stopSign.Deal(mouduleCode)
        return false
    }

    //标记请求; 如果是首次请求, 则自增请求数量, 否则啥也不干
    if _, loaded := schdl.urlMap.LoadOrStore(uurl, &basic.UrlInfo{
        Status: basic.URL_STATUS_DOWNLOADING,
        Ref:refUrl,
        Depth:req.Depth(),
    }); !loaded {
        atomic.AddUint64(&schdl.urlCnt, 1)
    }

    //请求入缓存
    schdl.requestCache.Put(req)
    log.Debug("Send the req to Cache: ", req.HttpReq().URL.String(), "  ",
        schdl.requestCache.Length(), schdl.requestCache.Capacity())

    return true
}

//对分析出来的请求做合法性校验，
// 合法返回true
// 不合法返回false
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

    //已经处理过的URL; 需要进一步判断不再处理
    v, ok := schdl.urlMap.Load(request.HttpReq().URL.String())
    if ok {
        //如果深度不匹配，则是非法的
        if v.(*basic.UrlInfo).Depth != request.Depth() {
            log.Debugf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", requestUrl)
            return false
        }
        //如果不是TimeOUT导致的重试，则也是非法的
        if v.(*basic.UrlInfo).Status != basic.URL_STATUS_HEAD_TIMEOUT && v.(*basic.UrlInfo).Status != basic.URL_STATUS_GET_TIMEOUT {
            log.Debugf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", requestUrl)
            return false
        }
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
