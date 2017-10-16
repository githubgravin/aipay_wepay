package base

import (
	"golib/gerror"
	"gopkg.in/ini.v1"
	"unGateWay/Services/pub"
)

type BaseCfg struct {
	pub.BaseResource
	LogName string
}

func (t *BaseCfg) InitCfg(cfg *ini.Section) gerror.IError {
	t.LogName = cfg.Key("LogName").String()
	if t.LogName == "" {
		return gerror.NewR(12001, nil, "日志名称不能为空", t.LogName)
	}
	t.SetLogger(t.LogName)
	return nil
}

func (t *BaseCfg) GetCurrLogName() string {
	return t.LogName
}
