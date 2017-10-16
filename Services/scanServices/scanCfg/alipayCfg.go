package scanCfg

import (
	"crypto/rsa"
	"encoding/json"
	"golib/gerror"
	"golib/security"
	"gopkg.in/ini.v1"
	"net"
	"unGateWay/Services/base"
)

type AlipayCfg struct {
	base.BaseCfg
	ScanCommCfg
	AppId                string
	SysServiceProviderID string
	InsIdCd              string
	LocalAddr            net.Addr
	NotiUrl              string
	NotiNum              int
	NotifyUrl            string
	RemoteURL            string
	RemoteFileURL        string
	ReqTimeOut           int
	QueryCnt             int
	QueryInt             int
	CancelCnt            int
	FilePath             string
	SignPrivKey          *rsa.PrivateKey
	EncPubKey            *rsa.PublicKey
	TranTimeOut          string
	OrderTimeOut         string
	PidMap               map[string]string
}

func (t *AlipayCfg) InitCfg(cfg *ini.Section) gerror.IError {
	var err error

	gerr := t.BaseCfg.InitCfg(cfg)
	if gerr != nil {
		return gerr
	}

	t.AppId = cfg.Key("APPID").String()
	t.SysServiceProviderID = cfg.Key("SysServiceProviderID").String()
	t.InsIdCd = cfg.Key("InsIdCd").String()

	ipr, err := net.ResolveIPAddr("ip", cfg.Key("LocalAddr").String())
	t.LocalAddr = &net.TCPAddr{IP: ipr.IP}
	t.NotiUrl = cfg.Key("NotiUrl").String()
	t.NotiNum = cfg.Key("NotiNum").MustInt(3)
	t.NotifyUrl = cfg.Key("NotifyUrl").String()
	t.RemoteURL = cfg.Key("RemoteURL").String()
	t.RemoteFileURL = cfg.Key("RemoteFileURL").String()
	t.ReqTimeOut = cfg.Key("ReqTimeOut").MustInt(60)
	t.QueryCnt = cfg.Key("QueryCnt").MustInt(5)
	t.QueryInt = cfg.Key("QueryInt").MustInt(2)
	t.CancelCnt = cfg.Key("CancelCnt").MustInt(2)
	t.FilePath = cfg.Key("FilePath").String()

	t.SignPrivKey, err = security.GetRsaPrivateKey(cfg.Key("SignKeyFile").String())
	if err != nil {
		return gerror.NewR(23010, err, "加载签名密钥失败")
	}
	t.EncPubKey, err = security.GetRsaPublicKey(cfg.Key("EncKeyFile").String())
	if err != nil {
		return gerror.NewR(23020, err, "加载证书失败")
	}

	t.TranTimeOut = cfg.Key("TranTimeOut").String()
	t.OrderTimeOut = cfg.Key("OrderTimeOut").String()

	pidMapStr := cfg.Key("PidMap").String()
	err = json.Unmarshal([]byte(pidMapStr), &t.PidMap)
	if err != nil {
		return gerror.NewR(23030, err, "加载PidMap失败", pidMapStr)
	}

	return nil
}

func (t *AlipayCfg) GetSvrPid(insId string) string {
	if t.PidMap == nil {
		return t.SysServiceProviderID
	}
	pid, ok := t.PidMap[insId]
	if !ok {
		return t.SysServiceProviderID
	}
	return pid
}
