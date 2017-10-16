package scanServices

import (
	"golib/gerror"
	"unGateWay/Services/scanServices/alipay"
	"unGateWay/Services/scanServices/scanCfg"
	"unGateWay/Services/scanServices/wepay"
)

type NotiServices struct {
	scanType string
	insId    string
	reqMsg   []byte
}

func NewNotiServices(scanType, insId string, reqMsg []byte) (*NotiServices, gerror.IError) {
	var gerr gerror.IError

	if !scanCfg.HasInit() {
		gerr = scanCfg.InitAllScanCfg()
		if gerr != nil {
			return nil, gerr
		}
	}

	notiSvr := &NotiServices{}
	notiSvr.scanType = scanType
	notiSvr.insId = insId
	notiSvr.reqMsg = reqMsg

	return notiSvr, nil
}

func (t *NotiServices) Run() ([]byte, gerror.IError) {

	var svr ITranServices
	var gerr gerror.IError

	cfg := scanCfg.GetInsScanCfg(t.insId)
	if cfg == nil {
		return nil, gerror.NewR(14001, nil, "未找到机构对应配置信息", t.insId)
	}

	switch t.scanType {
	case "W":
		svr, gerr = wepay.NewTranServices(cfg)
	case "A":
		svr, gerr = alipay.NewTranServices(cfg)
	default:
		return nil, gerror.NewR(14010, nil, "不支持的类型", t.scanType)
	}
	if gerr != nil {
		return nil, gerr
	}
	svr.SetLogger("Noti_" + svr.GetCurrLogName())
	res, gerr := svr.DoNotify(t.reqMsg)

	return []byte(res), nil
}
