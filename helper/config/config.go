package config

import (
    "strconv"
    "github.com/Unknwon/goconfig"
    "github.com/hq-cml/spider-go/basic"
)

//解析配置文件
func ParseConfig(confPath string) (*basic.SpiderConf, error){
    cfg, err := goconfig.LoadConfigFile(confPath)
    if err != nil {
        panic("Load conf file failed!")
    }

    c := &basic.SpiderConf{

    }
    v, err := cfg.GetValue("spider", "grabDepth")
    if err != nil {
        panic("Load conf ip failed!")
    }
    i, err := strconv.Atoi(v)
    if err != nil {
        panic("strconv.Atoi failed!")
    }
    c.GrabDepth = i

    return c, nil
}