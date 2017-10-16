package Config

import (
	"golib/gerror"
	"gopkg.in/ini.v1"
)

type ICfg interface {
	InitCfg(cfg *ini.Section) gerror.IError
}
