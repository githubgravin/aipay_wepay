package scanServices

import (
	"golib/gerror"
	"unGateWay/Services/scanServices/alipay"
	"unGateWay/Services/scanServices/scanCfg"
	"unGateWay/Services/scanServices/wepay"
)

type BusServices struct {
	BusType string
	insId   string
	reqMsg  []byte
}

func NewBusServices(scanType, insId string, reqMsg []byte) (*BusServices, gerror.IError) {
	var gerr gerror.IError

	if !scanCfg.HasInit() {
		gerr = scanCfg.InitAllScanCfg()
		if gerr != nil {

			return nil, gerr
		}
	}

	BusSvr := &BusServices{}
	BusSvr.BusType = scanType
	BusSvr.insId = insId
	BusSvr.reqMsg = reqMsg

	return BusSvr, nil
}

func (t *BusServices) Run() ([]byte, gerror.IError) {

	var svr ITranServices
	var gerr gerror.IError

	cfg := scanCfg.GetInsScanCfg(t.insId)
	if cfg == nil {
		return nil, gerror.NewR(14001, nil, "未找到机构对应配置信息", t.insId)
	}

	switch t.BusType {
	case "W":
		svr, gerr = wepay.NewTranServices(cfg)
	case "A":
		svr, gerr = alipay.NewTranServices(cfg)
	default:
		return nil, gerror.NewR(14010, nil, "不支持的类型", t.BusType)
	}
	if gerr != nil {
		return nil, gerr
	}
	svr.SetLogger("Bus_" + svr.GetCurrLogName())
	res, gerr := svr.DoBusSvr(t.reqMsg)
	if gerr != nil {
		return nil, gerr
	}

	return []byte(res), nil
}
