package HttpsSvr

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// Name for adapter with now support
const (
	AdaptQr     = "/qrpay"     //交易
	AdaptQrNoti = "/qrpayNoti" // 信息
	AdaptQrMcht = "/qrmcht"    //商户
	AdaptQrSett = "/qrsett"
	AdaptAgtPay = "/agtpay"
)

type HandleInfo struct {
	URL         string
	HttpHandler func(http.ResponseWriter, *http.Request, httprouter.Params)
	ReqType     string
}

//URL适配器
var adpHttpsHanders = make(map[string]*HandleInfo)
var adpHttpHanders = make(map[string]*HandleInfo)

type HttpType int

const (
	OnlyHttp HttpType = iota + 1
	OnlyHttps
	BothHttp
)

//注册函数
func RegisteAdpHander(adpName string, hdf *HandleInfo, httpType HttpType) {
	if hdf == nil {
		panic("HttpHandler: Register provide is nil")
	}
	switch httpType {
	case OnlyHttp:
		if _, dup := adpHttpHanders[adpName]; dup {
			panic("HttpHandler: Register called twice for provider " + adpName)
		}
		fmt.Println("注册http")
		adpHttpHanders[adpName] = hdf
	case OnlyHttps:
		if _, dup := adpHttpsHanders[adpName]; dup {
			panic("HttpHandler: Register called twice for provider " + adpName)
		}
		//fmt.Println(adpHttpsHanders[adpName].ReqType)
		//fmt.Println(adpHttpsHanders[adpName].ReqType)
		fmt.Println("注册https")
		adpHttpsHanders[adpName] = hdf
	case BothHttp:
		if _, dup := adpHttpHanders[adpName]; dup {
			panic("HttpHandler: Register called twice for provider " + adpName)
		}
		adpHttpHanders[adpName] = hdf
		if _, dup := adpHttpsHanders[adpName]; dup {
			panic("HttpHandler: Register called twice for provider " + adpName)
		}
		adpHttpsHanders[adpName] = hdf
	default:
		panic("非法的注册类型")
	}
}

/*根据配置信息匹配域名处理函数*/
func GetHandleInfo(adpName string, httpType HttpType) (string,
	func(http.ResponseWriter, *http.Request, httprouter.Params)) {

	var adpHand map[string]*HandleInfo
	switch httpType {
	case OnlyHttp:
		adpHand = adpHttpHanders
	case OnlyHttps:
		adpHand = adpHttpsHanders
	default:
		panic("非法的http类型")
	}
	ah, ok := adpHand[adpName]
	if !ok {
		panic("HttpHandler: 未注册HTTP处理器:" + adpName)
	}

	return ah.URL, ah.HttpHandler
}
