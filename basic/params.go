package basic

import (
	"errors"
	"fmt"
)

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

//池基本参数容器的描述
func (p *PoolParams) String() string {
	if p.description == "" {
		p.description = fmt.Sprintf("{ pageDownloaderPoolSize: %d, analyzerPoolSize: %d }", p.downloaderPoolSize, p.analyzerPoolSize)
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
