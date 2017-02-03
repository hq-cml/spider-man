package downloader
/*
 * 下载器池
 */

//池类型接口
type PageDownloaderPoolIntfs interface {
    Get() (PageDownloaderIntfs, error) // 从池中获取一个下载器
    Put(dl PageDownloaderIntfs) error // 归还一个下载器到池子中
    Total() uint32 //获得池子总容量
    Used() uint32 //获得正在被使用的下载器数量
}