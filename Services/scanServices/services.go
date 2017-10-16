package scanServices

import (
	"golib/gerror"
	"unGateWay/Services/scanServices/alipay"
	"unGateWay/Services/scanServices/scanCfg"
	"unGateWay/Services/scanServices/scanModel"
	"unGateWay/Services/scanServices/wepay"
)

type ScanServices struct {
	req *scanModel.TransMessage
	rsp *scanModel.TransMessage
}

func NewScanServices(reqMsg []byte) (*ScanServices, gerror.IError) {
	var gerr gerror.IError

	if !scanCfg.HasInit() {
		gerr = scanCfg.InitAllScanCfg()
		if gerr != nil {
			return nil, gerr
		}
	}

	scanSvr := &ScanServices{}
	scanSvr.req, gerr = scanModel.UnPackReq(reqMsg)

	if gerr != nil {
		return nil, gerr
	}
	return scanSvr, nil
}

func (t *ScanServices) Run() ([]byte, gerror.IError) {

	var svr ITranServices
	var gerr gerror.IError

	cfg := scanCfg.GetInsScanCfg(t.req.MsgBody.ChnInsIdCd)
	if cfg == nil {
		return nil, gerror.NewR(14001, nil, "未找到机构对应配置信息", t.req.MsgBody.ChnInsIdCd)
	}

	switch t.req.MsgBody.Biz_cd {
	case "0000007":
		svr, gerr = wepay.NewTranServices(cfg)
	case "0000008":
		svr, gerr = alipay.NewTranServices(cfg)
	default:
		return nil, gerror.NewR(14005, nil, "无法识别的业务类型", t.req.MsgBody.Biz_cd)
	}

	if gerr != nil {
		return nil, gerr
	}
	svr.SetLogger("Trn_" + svr.GetCurrLogName())
	t.rsp, gerr = svr.DoServices(t.req)
	if gerr != nil {
		return nil, gerr
	}

	res, err := t.rsp.PackRsp()
	if err != nil {
		return nil, gerror.NewR(14010, err, "生成应答报文失败")
	}

	//return nil, gerror.NewR(15001, nil, "模拟超时")

	return res, nil
}
