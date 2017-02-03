package downloader

import "github.com/hq-cml/spider-go/basic"

/*
 * 网页下载器存在于下载器池中，每个下载器有自己的Id
 *
 */

type PageDownloaderIntfs interface {
    Id() uint32 //获得下载器的Id
    Download(req basic.Request) (*basic.Response, error) //实际的下载行为
}
