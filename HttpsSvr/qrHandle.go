package HttpsSvr

import (
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"unGateWay/Services/scanServices"
)

func init() {
	//域名处理函数注册
	hdInfo := HandleInfo{URL: AdaptQr, HttpHandler: ProdQrProc, ReqType: http.MethodPost}
	RegisteAdpHander(AdaptQr, &hdInfo, OnlyHttps)
	notiInfo := HandleInfo{URL: AdaptQrNoti + "/:scanType/:insId", HttpHandler: NotiQrProc, ReqType: http.MethodPost}
	RegisteAdpHander(AdaptQrNoti, &notiInfo, OnlyHttp)
	busInfo := HandleInfo{URL: AdaptQrMcht + "/:busType/:insId", HttpHandler: BusProc, ReqType: http.MethodGet}
	RegisteAdpHander(AdaptQrMcht, &busInfo, OnlyHttp)
}

/*扫码信息解析*/
func ProdQrProc(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	msgBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		AppInst.Error("读取请求报文体失败", err)
		return
	}
	AppInst.Debug("收到请求报文", string(msgBody)) //111111111

	svr, gerr := scanServices.NewScanServices(msgBody)
	if gerr != nil {
		w.WriteHeader(500)
		AppInst.Error("创建服务失败", gerr)
		return
	}

	resp, gerr := svr.Run()
	if gerr != nil {
		w.WriteHeader(500)
		AppInst.Error("调用了服务失败", gerr)
		return
	}

	w.Write(resp)
	AppInst.Debug("发送应答成功", string(resp))

	return
}

func NotiQrProc(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	scanType := ps.ByName("scanType")
	insId := ps.ByName("insId")

	msgBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		AppInst.Error("读取异步通知失败", err)
		w.WriteHeader(500)
		return
	}
	AppInst.Debug("收到异步通知", string(msgBody), scanType, insId)

	svr, gerr := scanServices.NewNotiServices(scanType, insId, msgBody)
	if gerr != nil {
		AppInst.Error("异步服务失败", gerr)
		return
	}
	res, gerr := svr.Run()
	if gerr != nil {
		AppInst.Error("处理异步通知失败", gerr)
		w.WriteHeader(500)
		return
	}
	w.Write(res)

	AppInst.Debug("异步通知处理成功", string(res))
	return
}

func BusProc(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	busType := ps.ByName("busType")
	insId := ps.ByName("insId")

	AppInst.Debug("收到业务请求", r.URL.RawQuery, busType, insId)

	svr, gerr := scanServices.NewBusServices(busType, insId, []byte(r.URL.RawQuery))
	if gerr != nil {
		AppInst.Error("业务服务失败", gerr)
		return
	}
	res, gerr := svr.Run()
	if gerr != nil {
		AppInst.Error("处理业务请求失败", gerr)
		w.WriteHeader(500)
		return
	}
	w.Write(res)

	AppInst.Debug("业务请求处理成功", string(res))
	return
}
