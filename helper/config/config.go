package config

import (
	"github.com/Unknwon/goconfig"
	"github.com/hq-cml/spider-go/basic"
)

//解析配置文件
func ParseConfig(confPath string) (*basic.SpiderConf, error) {
	cfg, err := goconfig.LoadConfigFile(confPath)
	if err != nil {
		panic("Load conf file failed!")
	}

	c := &basic.SpiderConf{}
	c.GrabDepth, err = cfg.Int("spider", "grabDepth")
	if err != nil {
		panic("Load conf grabDepth failed!")
	}

	c.PluginKey, err = cfg.GetValue("spider", "pluginKey")
	if err != nil {
		panic("Load conf pluginKey failed!")
	}

	c.RequestChanCapcity, err = cfg.Int("spider", "requestChanCapcity")
	if err != nil {
		panic("Load conf requestChanCapcity failed!")
	}

	c.ResponseChanCapcity, err = cfg.Int("spider", "responseChanCapcity")
	if err != nil {
		panic("Load conf responseChanCapcity failed!")
	}

	c.EntryChanCapcity, err = cfg.Int("spider", "entryChanCapcity")
	if err != nil {
		panic("Load conf entryChanCapcity failed!")
	}

	c.ErrorChanCapcity, err = cfg.Int("spider", "errorChanCapcity")
	if err != nil {
		panic("Load conf errorChanCapcity failed!")
	}

	c.DownloaderPoolSize, err = cfg.Int("spider", "downloaderPoolSize")
	if err != nil {
		panic("Load conf downloaderPoolSize failed!")
	}

	c.AnalyzerPoolSize, err = cfg.Int("spider", "analyzerPoolSize")
	if err != nil {
		panic("Load conf analyzerPoolSize failed!")
	}

	c.MaxIdleCount, err = cfg.Int("spider", "maxIdleCount")
	if err != nil {
		panic("Load conf maxIdleCount failed!")
	}

	c.IntervalNs, err = cfg.Int("spider", "intervalNs")
	if err != nil {
		panic("Load conf intervalNs failed!")
	}

	c.SummaryDetail, err = cfg.Bool("spider", "summaryDetail")
	if err != nil {
		panic("Load conf summaryDetail failed!" + err.Error())
	}

	c.SummaryInterval, err = cfg.Int("spider", "summaryInterval")
	if err != nil {
		panic("Load conf summaryInterval failed!")
	}

	return c, nil
}
