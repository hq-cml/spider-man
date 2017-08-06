package downloader

import (
    "github.com/hq-cml/spider-go/middleware/pool"
    "fmt"
    "reflect"
    "errors"
)


/********************************下载器池********************************/
//New,创建网页下载器
func NewDownloaderPool(total uint32, gen GenDownloaderFunc) (DownloaderPoolIntfs, error) {
    //直接调用gen()，利用反射获取期类型
    etype := reflect.TypeOf(gen())

    //gen()的返回值是DownloaderIntfs，但是它包含了pool.EntityIntfs所有声明方法
    //所以可以认为DownloaderIntfs是pool.EntityIntfs
    genEntity := func() pool.EntityIntfs {
        return gen()
    }
    pool, err := pool.NewPool(total, etype, genEntity)
    if err != nil {
        return nil, err
    }

    dl := &DownloaderPool{
        pool : pool,
        etype: etype,
    }

    return dl, nil
}

//*DownloaderPool实现DownloaderPoolIntfs
func (dlpool *DownloaderPool) Get() (DownloaderIntfs, error) {
    entity, err := dlpool.pool.Get()
    if err != nil {
        return nil, err
    }
    dl, ok := entity.(DownloaderIntfs)
    if !ok {
        msg := fmt.Sprintf("The type of entity is not %s!\n", dlpool.etype)
        panic(errors.New(msg))
    }

    return dl, nil
}

func (dlpool *DownloaderPool) Put(dl DownloaderIntfs) error {
    return dlpool.pool.Put(dl)
}

func (dlpool *DownloaderPool) Total() uint32 {
    return dlpool.pool.Total()
}

func (dlpool *DownloaderPool) Used() uint32 {
    return dlpool.pool.Used()
}
