package scanServices

import (
	"golib/gerror"
	"unGateWay/Services/scanServices/scanModel"
)

type ITranServices interface {
	GetCurrLogName() string
	SetLogger(logName string)
	DoServices(req *scanModel.TransMessage) (*scanModel.TransMessage, gerror.IError)
	DoNotify(req []byte) ([]byte, gerror.IError)
	DoBusSvr(req []byte) ([]byte, gerror.IError)
}
