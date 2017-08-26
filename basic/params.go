package basic

import (
	"errors"
	"fmt"
)

//通道参数的容器的描述模板。
var channelParamsTemplate string = "{ reqChanLen: %d, respChanLen: %d, entryChanLen: %d, errorChanLen: %d }"

//池基本参数容器的描述模板。
var poolParamsTemplate string = "{ pageDownloaderPoolSize: %d, analyzerPoolSize: %d }"

//创建通道参数的容器。
func NewChannelParams(reqChanLen uint, respChanLen uint, entryChanLen uint, errorChanLen uint) ChannelParams {
	return ChannelParams{
		reqChanLen:   reqChanLen,
		respChanLen:  respChanLen,
		entryChanLen: entryChanLen,
		errorChanLen: errorChanLen,
	}
}

func (p *ChannelParams) Check() error {
	if p.reqChanLen == 0 {
		return errors.New("The request channel max length (capacity) can not be 0!\n")
	}
	if p.respChanLen == 0 {
		return errors.New("The response channel max length (capacity) can not be 0!\n")
	}
	if p.entryChanLen == 0 {
		return errors.New("The entry channel max length (capacity) can not be 0!\n")
	}
	if p.errorChanLen == 0 {
		return errors.New("The error channel max length (capacity) can not be 0!\n")
	}
	return nil
}

func (args *ChannelParams) String() string {
	if args.description == "" {
		args.description = fmt.Sprintf(channelParamsTemplate, args.reqChanLen, args.respChanLen,
			args.entryChanLen, args.errorChanLen)
	}
	return args.description
}

// 获得请求通道的长度。
func (p *ChannelParams) ReqChanLen() uint {
	return p.reqChanLen
}

// 获得响应通道的长度。
func (p *ChannelParams) RespChanLen() uint {
	return p.respChanLen
}

// 获得条目通道的长度。
func (p *ChannelParams) EntryChanLen() uint {
	return p.entryChanLen
}

// 获得错误通道的长度。
func (p *ChannelParams) ErrorChanLen() uint {
	return p.errorChanLen
}

// 创建池基本参数的容器。
func NewPoolParams(dlSize uint32, anlSize uint32) PoolParams {
	return PoolParams{
		downloaderPoolSize: dlSize,
		analyzerPoolSize:   anlSize,
	}
}

func (p *PoolParams) Check() error {
	if p.downloaderPoolSize == 0 {
		return errors.New("The downloader pool size can not be 0!\n")
	}
	if p.analyzerPoolSize == 0 {
		return errors.New("The analyzer pool size can not be 0!\n")
	}
	return nil
}

func (p *PoolParams) String() string {
	if p.description == "" {
		p.description = fmt.Sprintf(poolParamsTemplate, p.downloaderPoolSize, p.analyzerPoolSize)
	}
	return p.description
}

// 获得网页下载器池的尺寸。
func (args *PoolParams) DownloaderPoolSize() uint32 {
	return args.downloaderPoolSize
}

// 获得分析器池的尺寸。
func (args *PoolParams) AnalyzerPoolSize() uint32 {
	return args.analyzerPoolSize
}
