package downloader

import (
	"github.com/hq-cml/spider-go/basic"
	"github.com/hq-cml/spider-go/helper/idgen"
	//"github.com/hq-cml/spider-go/middleware/pool"
	"net/http"
	//"reflect"
	"io/ioutil"
	"bytes"
	"github.com/hq-cml/spider-go/helper/log"
	"errors"
	"fmt"
)

/***********************************下载器**********************************/
//网页下载器，*Downloader实现SpiderEntity接口
type Downloader struct {
	id         uint64 //ID
	httpClient http.Client
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
		httpClient: *client,
	}
}

//实际下载的工作，将http的返回结果，封装到basic.Response中
func (dl *Downloader) Download(req basic.Request) (*basic.Response, error) {
	httpReq := req.HttpReq()
	log.Infof("Download the request (reqUrl=%s)... Depth: (%d) \n",
		httpReq.URL.String(), req.Depth())

	httpResp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	//仅支持返回码200的响应
	if httpResp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Unsupported status code %d.", httpResp.StatusCode))
		return nil, err
	}

	//这个地方是一个go处理response的套路，读取了http.responseBody之后，如果不做处理则再次ReadAll的时候将出现空
	//Body内部有读取位置指针，一般的处理都是先close掉真实的body（释放连接），然后在利用NopCloser封装
	//一个伪造的ReaderCloser接口变量，然后赋值给Body，此时的Body已经篡改，但是这应该不会有什么问题
	//因为主要就是Body本身也是ReaderCloser实现类型，就只有read和close操作
	p, _ := ioutil.ReadAll(httpResp.Body)
	httpResp.Body.Close()
	httpResp.Body = ioutil.NopCloser(bytes.NewBuffer(p))

	return basic.NewResponse(httpResp, req.Depth()), nil
}
