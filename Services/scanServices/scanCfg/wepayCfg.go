package scanCfg

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"golib/gerror"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"os"
	"unGateWay/Services/base"
)

type WepayCfg struct {
	base.BaseCfg
	ScanCommCfg
	ServerName      string
	ServerId        string
	AppId           string
	WemchtId        string
	AppSecret       string
	BillDir         string
	OutIp           string
	NotiUrl         string
	NotiNum         int
	NotifyUrl       string
	RemoteUrl       string
	BindAddr        string
	PrivateCertFile string
	PrivateKeyFile  string
	PrivateCert     tls.Certificate
	PrivateKey      *rsa.PrivateKey
	ServerTimeOut   int
	QueryCnt        int
	QueryInt        int
	WeChnCfg        map[string]WepahChannCfg
	DefChnId        string
}

type WepahChannCfg struct {
	JsapiPath      string `json:"jsapi_path,omitempty"`
	SubAppid       string `json:"sub_appid,omitempty"`
	SubscribeAppid string `json:"subscribe_appid,omitempty"`
}

func (t *WepayCfg) InitCfg(cfg *ini.Section) gerror.IError {

	t.BaseCfg.InitCfg(cfg)

	t.ServerName = cfg.Key("servername").String()
	t.ServerId = cfg.Key("serverid").String()
	t.AppId = cfg.Key("appid").String()
	t.WemchtId = cfg.Key("wemchtid").String()
	t.AppSecret = cfg.Key("appsecret").String()
	t.BillDir = cfg.Key("billdir").String()
	t.OutIp = cfg.Key("outip").String()
	t.NotiUrl = cfg.Key("NotiUrl").String()
	t.NotiNum = cfg.Key("NotiNum").MustInt(3)
	t.NotifyUrl = cfg.Key("NotifyUrl").String()
	t.RemoteUrl = cfg.Key("remoteurl").String()
	t.BindAddr = cfg.Key("BindAddr").String()
	t.PrivateCertFile = cfg.Key("PrivateCert").String()
	t.PrivateKeyFile = cfg.Key("PrivateKey").String()
	t.ServerTimeOut = cfg.Key("ServerTimeOut").MustInt(50)
	t.QueryCnt = cfg.Key("QueryCnt").MustInt(0)
	t.QueryInt = cfg.Key("QueryInt").MustInt(0)

	var err error
	t.PrivateCert, err = tls.LoadX509KeyPair(t.PrivateCertFile, t.PrivateKeyFile)
	if err != nil {
		return gerror.NewR(13000, err, "加载密钥失败")
	}

	if t.QueryCnt == 0 || t.QueryInt == 0 {
		return gerror.NewR(13001, nil, "非法的配置参数", t)
	}

	//加载渠道商配置
	t.DefChnId = cfg.Key("defChnId").String()
	chnCfg := cfg.Key("WepChnCfg").String()
	if chnCfg != "" {
		f, err := os.Open(chnCfg)
		if err != nil {
			return gerror.NewR(13005, err, "打开配置文件失败", chnCfg)
		}
		tch, err := ioutil.ReadAll(f)
		if err != nil {
			return gerror.NewR(13006, err, "读配置文件失败", chnCfg)
		}
		t.WeChnCfg = make(map[string]WepahChannCfg, 0)
		err = json.Unmarshal(tch, &t.WeChnCfg)
		if err != nil {
			return gerror.NewR(13010, err, "加载渠道配置失败")
		}
	}

	return nil
}
