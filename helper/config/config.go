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
	if c.GrabMaxDepth, err = cfg.Int("spider", "grabMaxDepth"); err != nil {
		panic("Load conf grabDepth failed!")
	}

	if c.RequestChanCapcity, err = cfg.Int("spider", "requestChanCapcity"); err != nil {
		panic("Load conf requestChanCapcity failed!")
	}

	if c.ResponseChanCapcity, err = cfg.Int("spider", "responseChanCapcity"); err != nil {
		panic("Load conf responseChanCapcity failed!")
	}

	if c.ItemChanCapcity, err = cfg.Int("spider", "itemChanCapcity"); err != nil {
		panic("Load conf itemChanCapcity failed!")
	}

	if c.ErrorChanCapcity, err = cfg.Int("spider", "errorChanCapcity"); err != nil {
		panic("Load conf errorChanCapcity failed!")
	}

	if c.DownloaderPoolSize, err = cfg.Int("spider", "downloaderPoolSize"); err != nil {
		panic("Load conf downloaderPoolSize failed!")
	}

	if c.AnalyzerPoolSize, err = cfg.Int("spider", "analyzerPoolSize"); err != nil {
		panic("Load conf analyzerPoolSize failed!")
	}

	if c.MaxIdleCount, err = cfg.Int("spider", "maxIdleCount"); err != nil {
		panic("Load conf maxIdleCount failed!")
	}

	if c.IntervalNs, err = cfg.Int("spider", "intervalNs"); err != nil {
		panic("Load conf intervalNs failed!")
	}

	if c.SummaryDetail, err = cfg.Bool("spider", "summaryDetail"); err != nil {
		panic("Load conf summaryDetail failed!" + err.Error())
	}

	if c.CrossSite, err = cfg.Bool("spider", "crossSite"); err != nil {
		panic("Load conf crossSite failed!" + err.Error())
	}

	if c.SummaryInterval, err = cfg.Int("spider", "summaryInterval"); err != nil {
		panic("Load conf summaryInterval failed!")
	}

	if c.PluginKey, err = cfg.GetValue("plugin", "pluginKey"); err != nil {
		panic("Load conf pluginKey failed!")
	}

	if c.LogPath, err = cfg.GetValue("log", "logPath"); err != nil {
		panic("Load conf logPath failed!")
	}

	if c.LogLevel, err = cfg.GetValue("log", "logLevel"); err != nil {
		panic("Load conf logLevel failed!")
	}

	if c.Pprof, err = cfg.Bool("pprof", "pprof"); err != nil {
		panic("Load conf pprof failed!" + err.Error())
	}

	if c.PprofPort, err = cfg.GetValue("pprof", "pprofPort"); err != nil {
		panic("Load conf maxIdleCount failed!")
	}

	if c.Step, err = cfg.Bool("debug", "step"); err != nil {
		panic("Load conf step failed!" + err.Error())
	}

	if c.SkipBinFile, err = cfg.Bool("skip", "skipBinFile"); err != nil {
		panic("Load conf skipBinFile failed!" + err.Error())
	}

	return c, nil
}
