package scanCfg

import (
	"golib/gerror"
	"gopkg.in/ini.v1"
	"strings"
	"unGateWay/Config"
)

var (
	g_scancfg map[string]Config.ICfg //key:insid
	g_hasInit bool
)

type ScanCommCfg struct {
}

func InitAllScanCfg() gerror.IError {

	//初始化扫码配置信息
	fileName := Config.GetIniCfg().Section("").Key("ScanCfg").String()

	f, err := ini.Load(fileName)
	if err != nil {
		return gerror.NewR(13010, err, "加载scan配置失败")
	}

	g_scancfg = make(map[string]Config.ICfg, 0)

	insList := f.Section("").Key("InsList").String()
	insArr := strings.Split(insList, ",")

	for idx := range insArr {
		ins := insArr[idx]
		if len(ins) < 1 {
			return gerror.NewR(13020, nil, "非法的机构配置", ins, insArr)
		}
		var cfg Config.ICfg
		switch ins[0] {
		case 'W':
			cfg = &WepayCfg{}
		case 'A':
			cfg = &AlipayCfg{}
		}
		gerr := cfg.InitCfg(f.Section(ins))
		if gerr != nil {
			return gerr
		}
		g_scancfg[ins[1:]] = cfg
	}

	g_hasInit = true

	return nil
}

func GetInsScanCfg(insName string) Config.ICfg {
	return g_scancfg[insName]
}

func HasInit() bool {
	return g_hasInit
}
