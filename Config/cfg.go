package Config

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gogap/logrus_mate"
	"golib/gerror"
	"gopkg.in/ini.v1"
	"os"
)

var g_cfg *GlobalCfg

const DEFAULT = "default"

type GlobalCfg struct {
	cfg    *ini.File
	logCfg *logrus_mate.LogrusMate
}

func InitGlobalCfg(filename string) gerror.IError {

	var err error

	g_cfg = new(GlobalCfg)
	g_cfg.cfg, err = ini.Load(filename)
	if err != nil {
		return gerror.NewR(11001, err, "加载配置文件失败")
	}

	logFile := g_cfg.cfg.Section("").Key("logFile").String()
	if logFile == "" {
		return gerror.NewR(11010, nil, "日志文件未配置")
	}
	fmt.Printf("logFile :%s", logFile)
	runMode := g_cfg.cfg.Section("").Key("runMode").MustString("prod")
	os.Setenv("RUN_MODE", runMode)
	loggerConf, err := logrus_mate.LoadLogrusMateConfig(logFile)
	fmt.Println("\r\n.....%\r\n")
	fmt.Println(loggerConf)
	fmt.Println("\r\n.....%\r\n")
	if err != nil {
		return gerror.NewR(11020, err, "加载日志文件配置失败")
	}
	g_cfg.logCfg, err = logrus_mate.NewLogrusMate(loggerConf)
	fmt.Println(".....&\r\n")
	fmt.Println(g_cfg.logCfg)
	fmt.Println(".....&\r\n")
	if err != nil {
		return gerror.NewR(11030, err, "加载日志文件失败")
	}

	return nil

}

func GetNameLog(name string) *logrus.Logger {
	n := g_cfg.logCfg.Logger(name)
	if n == nil {
		n = g_cfg.logCfg.Logger(DEFAULT)
		if n == nil {
			n = g_cfg.logCfg.Logger()
		}
	}
	return n
}

func GetGlobalLog() *logrus.Logger {
	return g_cfg.logCfg.Logger(DEFAULT)
}

func GetIniCfg() *ini.File {
	return g_cfg.cfg
}
