package main

import (
    "net/http"
    "github.com/hq-cml/spider-go/logic/analyzer"
    "github.com/hq-cml/spider-go/logic/processchain"
    "github.com/hq-cml/spider-go/helper/log"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/logic/scheduler"
    "strings"
    "net/url"
    "fmt"
    "io"
    "errors"
    "github.com/PuerkitoBio/goquery"
    "time"
)

//生成HTTP客户端
func genHttpClient() *http.Client {
    return &http.Client{}
}

//获得响应解析函数的序列
func getResponseAnalysers() []analyzer.AnalyzeResponseFunc {
    analyzers := []analyzer.AnalyzeResponseFunc{
        parseForATag,
    }
    return analyzers
}

//响应解析函数。只解析“A”标签。
func parseForATag(httpResp *http.Response, grabDepth uint32) ([]basic.DataIntfs, []error) {
    //仅支持返回码200的响应
    if httpResp.StatusCode != 200 {
        err := errors.New(fmt.Sprintf("Unsupported status code %d. (httpResponse=%v)", httpResp))
        return nil, []error{err}
    }

    //对响应做一些处理
    var reqUrl *url.URL = httpResp.Request.URL     //记录下响应的请求（防止相对URL的问题）
    var httpRespBody io.ReadCloser = httpResp.Body
    defer func() {                                 //TODO 这一块应该放到框架中去？？？？
        if httpRespBody != nil {
            httpRespBody.Close()
        }
    }()


    dataList := make([]basic.DataIntfs, 0)
    errs := make([]error, 0)
    //开始解析
    doc, err := goquery.NewDocumentFromReader(httpRespBody)
    if err != nil {
        errs = append(errs, err)
        return dataList, errs
    }

    //查找“A”标签并提取链接地址
    doc.Find("a").Each(func(index int, sel *goquery.Selection) {
        href, exists := sel.Attr("href")
        // 前期过滤
        if !exists || href == "" || href == "#" || href == "/" {
            return
        }
        href = strings.TrimSpace(href)
        lowerHref := strings.ToLower(href)
        // 暂不支持对Javascript代码的解析。
        if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
            aUrl, err := url.Parse(href)
            if err != nil {
                errs = append(errs, err)
                return
            }
            //保证是绝对URL(如果当前是相对URL，则将当前URL拼接到主URL上，保证是绝对URL)
            if !aUrl.IsAbs() {
                aUrl = reqUrl.ResolveReference(aUrl)
            }
            httpReq, err := http.NewRequest("GET", aUrl.String(), nil)
            if err != nil {
                errs = append(errs, err)
            } else {
                req := basic.NewRequest(httpReq, grabDepth)
                //将新分析出来的请求，放入dataList
                dataList = append(dataList, req)
            }
        }
        text := strings.TrimSpace(sel.Text())
        //将新分析出来的Entry，放入dataList
        if text != "" {
            imap := make(map[string]interface{})
            imap["parent_url"] = reqUrl
            imap["a.text"] = text
            imap["a.index"] = index
            entry := basic.Entry(imap)
            dataList = append(dataList, &entry)
        }
    })
    return dataList, errs
}

// 获得条目处理链的序列。
func getEntryProcessors() []processchain.ProcessEntryFunc {
    entryProcessors := []processchain.ProcessEntryFunc {
        processEntry,
    }
    return entryProcessors
}

// 条目处理器。
func processEntry(entry basic.Entry) (result basic.Entry, err error) {
    if entry == nil {
        return nil, errors.New("Invalid entry!")
    }

    // 生成结果
    result = make(map[string]interface{})
    for k, v := range entry {
        result[k] = v
    }
    if _, ok := result["number"]; !ok {
        result["number"] = len(result)
    }
    //延时查看效果
    time.Sleep(10 * time.Millisecond)
    return result, nil
}


func main() {
    channelParams := basic.NewChannelParams(10, 10, 10, 10)
    poolParams := basic.NewPoolParams(3, 3)
    grabDepth := uint32(1)
    //httpClientGenerator := genHttpClient
    responseAnalysers := getResponseAnalysers()
    entryProcessors := getEntryProcessors()

    startUrl := "http://www.sogou.com"
    firstHttpReq , err := http.NewRequest("GET", startUrl, nil)
    if err != nil {
        log.Warnln(err.Error())
        return
    }

    // 创建调度器
    scheduler := scheduler.NewScheduler()
    scheduler.Start(channelParams, poolParams, grabDepth, genHttpClient,
        responseAnalysers, entryProcessors, firstHttpReq)


}
