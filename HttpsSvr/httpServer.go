package HttpsSvr

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"time"
	"unGateWay/Config"
	"unGateWay/Services/pub"
	"unGateWay/util"
)

type HttpSvrCfg struct {
	CertFile     string
	KeyFile      string
	IP           string
	Port         int
	ReadTimeOut  int
	WriteTimeOut int
}

/*HTTPS Server APP*/
type App struct {
	pub.BaseResource //继承基本资源
	Server           *http.Server
}

var AppInst *App

func InitModule(section string) error { //SECTION为HTTP或者HTTPS
	var isHttps bool

	cfg, err := Config.GetIniCfg().GetSection(section)
	if err != nil {
		return err
	}
	hCfg := new(HttpSvrCfg) //hcfg 为http配置信息
	hCfg.CertFile = cfg.Key("HttpsCertFile").String()
	hCfg.KeyFile = cfg.Key("HttpsKeyFile").String()
	hCfg.IP = cfg.Key("HttpsIP").MustString("127.0.0.1")
	hCfg.Port = cfg.Key("HttpsPort").MustInt(443)
	hCfg.ReadTimeOut = cfg.Key("ReadTimeOut").MustInt(10)
	hCfg.WriteTimeOut = cfg.Key("WriteTimeOut").MustInt(10)

	app := new(App)
	AppInst = app
	//fmt.Println(section)
	app.SetLogger(section)

	if hCfg.CertFile == "" || !util.FileExist(hCfg.CertFile) ||
		hCfg.KeyFile == "" || !util.FileExist(hCfg.KeyFile) {
		AppInst.Infof("[HttpsSvr] 密钥文件[%s] 证书文件[%s] 未配置;",
			hCfg.CertFile, hCfg.KeyFile)
		isHttps = false
	} else {
		isHttps = true
	}

	//启动HTTP Server
	Svr := new(http.Server)
	Svr.Addr = fmt.Sprintf("%s:%d", hCfg.IP, hCfg.Port)
	Svr.ReadTimeout = time.Duration(hCfg.ReadTimeOut) * time.Second
	Svr.WriteTimeout = time.Duration(hCfg.WriteTimeOut) * time.Second
	Svr.ErrorLog = log.New(Config.GetNameLog(section).Writer(), "", log.LstdFlags)

	AppInst.Debug("配置信息:", *hCfg)

	/*路由规则:根据产品参数配置*/
	router := httprouter.New()

	var adpHand map[string]*HandleInfo
	if isHttps {
		adpHand = adpHttpsHanders
	} else {
		adpHand = adpHttpHanders
	}
	for _, v := range adpHand {
		switch v.ReqType {
		case http.MethodPost:
			router.POST(v.URL, v.HttpHandler)
		case http.MethodGet:
			router.GET(v.URL, v.HttpHandler)
		default:
			app.Panic("不支持的类型", v.URL, v.ReqType)
		}
	}

	app.Server = Svr
	app.Server.Handler = router

	if isHttps {
		go app.Server.ListenAndServeTLS(hCfg.CertFile, hCfg.KeyFile)
	} else {
		go app.Server.ListenAndServe()
		fmt.Println("监听")
	}

	app.Info(section, "模块加载成功")
	return nil
}
