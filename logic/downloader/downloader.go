package downloader

import (
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/idgen"
	"net/http"
	"io/ioutil"
	"github.com/hq-cml/spider-go/helper/log"
	"errors"
	"fmt"
	"strings"
	"time"
)

/***********************************下载器**********************************/
//网页下载器，*Downloader实现SpiderEntity接口
type Downloader struct {
	id         uint64 //ID
	httpClient *http.Client
}
func (dl *Downloader) Id() uint64 {
	return dl.id
}

//下载器专用的id生成器
var downloaderIdGenerator *idgen.IdGenerator = idgen.NewIdGenerator()

//New
func NewDownloader(client *http.Client) *Downloader {
	id := downloaderIdGenerator.GetId()

	if client == nil {
		fmt.Println("BBBBBBBBBBBBBBBBBB----------")
		client = &http.Client{}
	}

	return &Downloader{
		id:         id,
		httpClient: client,
	}
}

//实际下载的工作，将http的返回结果，封装到basic.Response中
//bool返回值表示请求是否被skip
func (dl *Downloader) Download(req *basic.Request) (*basic.Response, bool, string, error) {
	httpReq := req.HttpReq()
	log.Infof("Start to Download the request (reqUrl=%s)... Depth: (%d) \n",
		httpReq.URL.String(), req.Depth())

	//跳过二进制文件下载
	if basic.Conf.SkipBinFile {
		skip, msg, err := dl.skipBinFile(req)
		if err != nil {
			log.Info("AAAAAAAAAAAAAAAAA----", errors.New("SkipBinFile("+ httpReq.URL.String() +") Error:" + err.Error()))
			if strings.Contains(strings.ToLower(err.Error()), "timeout") {
				return nil, false, "head timeout", errors.New("SkipBinFile("+ httpReq.URL.String() +") Error:" + err.Error())
			}
			return nil, false, "", errors.New("SkipBinFile("+ httpReq.URL.String() +") Error:" + err.Error())
		}

		if skip {
			return nil, true, msg, nil
		}
	}

	httpResp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "timeout") {
			return nil, false, "get timeout", errors.New("Download("+ httpReq.URL.String() +") Error:" + err.Error())
		}
		return nil, false, "", errors.New("Download("+ httpReq.URL.String() +") Error:" + err.Error())
	}
	defer httpResp.Body.Close()

	//仅支持返回码200的响应
	if httpResp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Unsupported status code %d. (ReqUrl=%s)",
			httpResp.StatusCode, httpReq.URL.String()))
		return nil, false, "", err
	}

	//这个地方是一个go处理response的套路，读取了http.responseBody之后，如果不做处理则再次ReadAll的时候将出现空
	//Body内部有读取位置指针，一般的处理都是先close掉真实的body（释放连接），然后在利用NopCloser封装
	//一个伪造的ReaderCloser接口变量，然后赋值给Body，此时的Body已经篡改，但是这应该不会有什么问题
	//因为主要就是Body本身也是ReaderCloser实现类型，就只有read和close操作
	//p, _ := ioutil.ReadAll(httpResp.Body)
	//httpResp.Body.Close()
	//httpResp.Body = ioutil.NopCloser(bytes.NewBuffer(p))

	body, ok := getBodyTimeout(httpResp, 30 * time.Second)
	if !ok {
		//TODO 其实，这是一种有损策略，为了保证服务不被全部卡死，只能牺牲
		//后续考虑将这些请求重新扔回队列中去
		err := errors.New("Time out：("+ httpReq.URL.String() +")")
		return nil, false, "read timeout", err
	}

	return basic.NewResponse(body,
		req.Depth(),
		httpResp.Header.Get("content-type"),
		req.HttpReq().URL.String()), false, "", nil
}

//运行中发现, 深度加大或者downloader数加大, 会发生内存暴涨
//分析发现有很多二进制的文件下载,将内存撑爆了,爬虫暂时不支持二进制文件
//过滤方式简单粗暴:
// 先判断url扩展名, 静态文件直接略过
// 如果扩展名不明显, 那只能发送一次HEAD方法的请求了, 但是这会导致多一次请求
// 暂时没有更好的方法
func (dl *Downloader)skipBinFile(req *basic.Request) (bool, string, error) {
	url := req.HttpReq().URL.String()

	//先通过扩展名来判断
	tmp := strings.Split(url, ".")
	ext := tmp[len(tmp)-1]
	if _, ok := basic.SkipBinFileExt[ext]; ok {
		log.Infof("Skip Request(%s)... Depth:(%d). Reason: Ext Invalid \n", url, req.Depth())
		return true, "Ext Invalid", nil
	}

	//TODO 这个地方可以继续优化，减少消耗，比如分析ext校验逃过的url的特点
	//通过HEAD请求来判断
	httpReq, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return true, "", err
	}
	resp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		return true, "", err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	contentLength := resp.Header.Get("Content-Length")
	if !strings.Contains(contentType, "text/html") {
		log.Infof("Skip Request(%s)... Depth:(%d). Reason: Content-Type Invalid \n", url, req.Depth())
		return true,
			   "Content-Type Invalid:" + contentType + ". Content-Length:" + contentLength,
			   nil
	}
	return false, "", nil
}

//从http.Reosponse中取出Body，并且支持超时
//由于服务器端实现的差异，在高并发情况下，ioutil.ReadAll常会出现迷之超时（官方解释ReadAll必须遇到EOF或者error才能结束）
//所以必须设置超时时间，防止downloader全部给卡死了
//返回true表示正常取值，返回false表示出现超时
func getBodyTimeout(resp *http.Response, timeout time.Duration) ([]byte, bool){
	var body []byte
	c := make(chan []byte)
	go func() {
		b, _ := ioutil.ReadAll(resp.Body)
		c <- b
	}()

	select {
	case <- time.After(timeout):
		return nil, false
	case body = <- c:
	}

	return body, true
}