package scheduler

import (
    "fmt"
    "errors"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/log"
    "github.com/hq-cml/spider-go/logic/downloader"
    "sync/atomic"
)

/*
 * 激活下载器，开始下载，下载器是一个独立的的goroutine无限循环，从请求通道中获取请求
 * 每个请求都扔给独立的goroutine去处理，但是goroutine并不一定能够立刻开始下载工作
 * 同时能够执行下载的goroutine数量,受到下载器池容量的制约, 其他goroutine会阻塞
 *
 *
 * 对于池子的使用，Analyzer和Downloader略有不同:
 * Downloader将取令牌操作，放在新建goroutine之外，原因是大概率下，Request数都非常大，如果如果放在里面会导致产生大量的等待的goroutine
 *      （因为Request Chan有Request Cache保护，所以不会出现全局阻塞，所以这么处理大面积降低了goroutine数量
 *
 * Analizer则不同，Response Chan没有特殊的保护，所以不能让其满了进而阻塞全局。并且由于通常Analyze程序是CPU型（DownLoader是IO型）
 *      很快都能够结束，所以直接将取令牌操作放在了goroutine内部，只控制同时可执行Analyzer数量，而非同时运行Analyzer数量
 */
func (schdl *Scheduler) activateDownloaders() {
    //downloader是异步独立的goroutine
    go func() {
        //无限循环，从请求通道中获取请求，每个请求都扔给独立的goroutine去处理
        for {
            request, ok := schdl.getReqestChan().Get()
            if !ok {
                //通道已关闭
                break
            }
            //类型断言
            req, ok := request.(basic.Request)
            if !ok {
                continue
            }

            //下载器池中取令牌，如果申请不到，就会阻塞等待在此处~
            entity, err := schdl.getDownloaderPool().Get()
            if err != nil {
                //msg := fmt.Sprintf("Downloader pool error: %s", err)
                schdl.sendError(err, DOWNLOADER_CODE)
                return
            }

            //每个请求都交给一个独立的goroutine来处理
            go schdl.download(req, entity)
        }
    }()
}

/*
 * 实际下载工作，下载goroutine的逻辑
 * 但是全部下载goroutine是受到下载器池子的约束的
 */
func (schdl *Scheduler) download(request basic.Request, entity basic.SpiderEntity) {
    atomic.AddUint64(&schdl.downloaderCnt, 1)          //原子加1
    defer atomic.AddUint64(&schdl.downloaderCnt, ^uint64(0)) //原子减1

    //panic错误兜底
    defer func() {
        if p := recover(); p != nil {
            msg := fmt.Sprintf("Fatal Download Error: %s. (ReqUrl=%s)\n", p, request.HttpReq().URL.String())
            log.Err(msg)
        }
    }()

    //注册延时归还令牌
    defer func() {
        err := schdl.getDownloaderPool().Put(entity)
        if err != nil {
            //msg := fmt.Sprintf("Downloader pool error: %s", err)
            schdl.sendError(err, DOWNLOADER_CODE)
        }
    }()

    //断言转换
    dl, ok := entity.(*downloader.Downloader)
    if !ok {
        msg := fmt.Sprint("Downloader pool Wrong type")
        schdl.sendError(errors.New(msg), DOWNLOADER_CODE)
        return
    }

    //实施下载
    v, ok := schdl.urlMap.Load(request.HttpReq().URL.String());
    if !ok {
        msg := fmt.Sprint("Can't find the url in urlMap:" + request.HttpReq().URL.String())
        schdl.sendError(errors.New(msg), DOWNLOADER_CODE)
        return
    }
    pInfo := v.(*basic.UrlInfo)

    moudleCode := generateModuleCode(DOWNLOADER_CODE, dl.Id())
    response, skip, msg, err := dl.Download(&request)
    if err != nil {
        //如果是HEAD或者GET请求超时, 且未达到最大重试次数, 那么进行重试
        if msg == "head timeout" {
            pInfo.Status = basic.URL_STATUS_HEAD_TIMEOUT
            if pInfo.Retry < basic.Conf.RetryTimes {
                schdl.sendRequestToCache(&request, moudleCode, pInfo.Ref)
                pInfo.Retry ++
                log.Warnf("Retry HEAD(%d): %s\n" , pInfo.Retry, request.HttpReq().URL.String())
                return
            }
        } else if msg == "get timeout" {
            pInfo.Status = basic.URL_STATUS_GET_TIMEOUT
            if pInfo.Retry < basic.Conf.RetryTimes {
                schdl.sendRequestToCache(&request, moudleCode, pInfo.Ref)
                pInfo.Retry ++
                log.Warnf("Retry GET(%d): %s\n" , pInfo.Retry, request.HttpReq().URL.String())
                return
            }
        } else if msg == "read timeout"{
            pInfo.Status = basic.URL_STATUS_READ_TIMEOUT
        } else {
            pInfo.Status = basic.URL_STATUS_FATAL_ERROR
        }
        pInfo.Msg = err.Error()
        err = errors.New("(URL:" + request.HttpReq().URL.String() + ") " + err.Error())
        schdl.sendError(err, moudleCode)
		return
    }

    //url标记成功
    if skip {
        pInfo.Status = basic.URL_STATUS_SKIP
        pInfo.Msg = msg
    } else {
        pInfo.Status = basic.URL_STATUS_DONE
    }

    //将resp放入
    if response != nil {
        schdl.sendToRespChan(*response, moudleCode)
    }
}

//获取Pool管理器持有的下载器Pool。
func (schdl *Scheduler) getDownloaderPool() basic.SpiderPool {
    p, err := schdl.poolManager.GetPool(DOWNLOADER_CODE)
    if err != nil {
        panic(err)
    }
    return p
}

//发送响应到通道管理器中的响应通道
func (schdl *Scheduler) sendToRespChan(resp basic.Response, mouduleCode string) bool {
    if schdl.stopSign.Signed() {
        //如果stop标记已经生效，则通道管理器可能已经关闭，此时不应该再进行通道写入
        schdl.stopSign.Deal(mouduleCode)
        return false
    }

    schdl.getResponseChan().Put(resp)
    return true
}