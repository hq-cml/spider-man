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
		client = &http.Client{}
	}

	return &Downloader{
		id:         id,
		httpClient: client,
	}
}

//实际下载的工作，将http的返回结果，封装到basic.Response中
//bool返回值表示请求是否被skip
func (dl *Downloader) Download(req basic.Request) (*basic.Response, bool, error) {
	httpReq := req.HttpReq()

	//跳过二进制文件下载
	if basic.Conf.SkipBinFile {
		skip, err := skipBinFile(&req)
		if err != nil {
			err := errors.New(fmt.Sprintf("skipBinFile Error: %s", err))
			return nil, false, err
		}

		if skip {
			return nil, true, nil
		}
	}

	log.Infof("Download the request (reqUrl=%s)... Depth: (%d) \n",
		httpReq.URL.String(), req.Depth())
	httpResp, err := dl.httpClient.Do(httpReq)   //TODO hang  9
	if err != nil {
		return nil, false,  err
	}
	defer httpResp.Body.Close()

	//仅支持返回码200的响应
	if httpResp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Unsupported status code %d.", httpResp.StatusCode))
		return nil, false, err
	}

	//这个地方是一个go处理response的套路，读取了http.responseBody之后，如果不做处理则再次ReadAll的时候将出现空
	//Body内部有读取位置指针，一般的处理都是先close掉真实的body（释放连接），然后在利用NopCloser封装
	//一个伪造的ReaderCloser接口变量，然后赋值给Body，此时的Body已经篡改，但是这应该不会有什么问题
	//因为主要就是Body本身也是ReaderCloser实现类型，就只有read和close操作
	//p, _ := ioutil.ReadAll(httpResp.Body)
	//httpResp.Body.Close()
	//httpResp.Body = ioutil.NopCloser(bytes.NewBuffer(p))


	body, _ := ioutil.ReadAll(httpResp.Body)    //TODO hang  9
	return basic.NewResponse(body,
		req.Depth(),
		httpResp.Header.Get("content-type"),
		req.HttpReq().URL.String()), false, nil
}

//运行中发现, 深度加大或者downloader数加大, 会发生内存暴涨
//分析发现有很多二进制的文件下载,将内存撑爆了,爬虫暂时不支持二进制文件
//过滤方式简单粗暴:
// 先判断url扩展名, 静态文件直接略过
// 如果扩展名不明显, 那只能发送一次HEAD方法的请求了, 但是这会导致多一次请求
// 暂时没有更好的方法
func skipBinFile(req *basic.Request) (bool, error) {
	url := req.HttpReq().URL.String()

	//先通过扩展名来判断
	tmp := strings.Split(url, ".")
	ext := tmp[len(tmp)-1]
	if _, ok := basic.SkipBinFileExt[ext]; ok {
		log.Infof("Skip Request(%s)... Depth:(%d). Reason: Ext Invalid \n", url, req.Depth())
		return true, nil
	}

	//通过HEAD请求来判断
	httpReq, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return true, err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return true, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		log.Infof("Skip Request(%s)... Depth:(%d). Reason: Content-Type Invalid \n", url, req.Depth())
		return true, nil
	}
	return false, nil
}