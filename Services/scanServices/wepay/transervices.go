package wepay

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"golib/gerror"
	"golib/modules/logr"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unGateWay/Config"
	"unGateWay/Services/scanServices/scanCfg"
	"unGateWay/Services/scanServices/scanModel"
)

type TranServices struct {
	TranCode    string
	URL         string
	WxJsStr     string
	RespCd      string /*交易应答码*/
	RespMsg     string /*交易应答信息*/
	OrigRespCd  string /*原交易应答码*/
	OrigRespMsg string /*原交易应答信息*/
	SystemFlag  bool   //系统接入标志
	ReqInfo     *Request
	RspInfo     *Response
	reqMsg      *scanModel.TransMessage
	*scanCfg.WepayCfg
}

func NewTranServices(cfg Config.ICfg) (*TranServices, gerror.IError) {
	var ok bool
	tranService := new(TranServices)

	tranService.WepayCfg, ok = cfg.(*scanCfg.WepayCfg)
	if !ok {
		return nil, gerror.NewR(12001, nil, "非法的配置信息", cfg)
	}

	tranService.ReqInfo = new(Request)
	tranService.RspInfo = new(Response)

	/*内部交易码*/
	tranService.ReqInfo.APPId = tranService.AppId
	tranService.ReqInfo.MchId = tranService.WemchtId

	return tranService, nil
}

func (tran *TranServices) SetWxOrderId(wxOrder string) {
	tran.ReqInfo.TransactionId = wxOrder
}

func (tran *TranServices) SetOrderId(orderId string) {
	tran.ReqInfo.OutTradeNo = orderId
}

func (tran *TranServices) SetRefundId(refundId string) {
	tran.ReqInfo.RefundId = refundId
}

func (tran *TranServices) SetMerId(merId string) {
	tran.ReqInfo.SubMchId = merId
}

func (tran *TranServices) SetBody(body string) {
	tran.ReqInfo.Body = body
}

func (tran *TranServices) SetAuthCode(authCode string) {
	tran.ReqInfo.AuthCode = authCode
}

/*设置交易金额*/
func (tran *TranServices) SetTxnAmt(txnAmt string) {
	tran.ReqInfo.TotalFee, _ = strconv.Atoi(txnAmt)
}

/*设置退款金额*/
func (tran *TranServices) SetRefundAmt(refundAmt string) {
	tran.ReqInfo.RefundFee, _ = strconv.Atoi(refundAmt)
}

/*设置操作员*/
func (tran *TranServices) SetOpUsrId(id string) {
	tran.ReqInfo.OpUserId = id
}

/*设置OpenId*/
func (tran *TranServices) SetOpenId(openid string) {
	tran.ReqInfo.OpenId = openid
}

/*设置SubOpenId*/
func (tran *TranServices) SetSubOpenId(sub_openid string) {
	tran.ReqInfo.SubOpenId = sub_openid
}

/*设置SubAPPId*/
func (tran *TranServices) SetSubAPPId(appId string) {
	tran.ReqInfo.SubAPPId = appId
}

/*设置MchName*/
func (tran *TranServices) SetMchName(mchName string) {
	tran.ReqInfo.MchName = mchName
}

/*设置 MchShtName*/
func (tran *TranServices) SetMchShtName(mchShortName string) {
	tran.ReqInfo.MchShtName = mchShortName
}

/*设置 ServicePhone*/
func (tran *TranServices) SetServicePhone(phone string) {
	tran.ReqInfo.ServicePhone = phone
}

/*设置 Business*/
func (tran *TranServices) SetBussiness(buss string) {
	tran.ReqInfo.Business = buss
}

/*设置 微信营销标志*/
func (tran *TranServices) SetGoodsTag(tag string) {
	tran.ReqInfo.GoodsTag = tag
}

/*设置 Business*/
func (tran *TranServices) SetMchRemark(remark string) {
	tran.ReqInfo.MchRemark = remark
}

/*设置 ChannelId*/
func (tran *TranServices) SetChannelId(channelId string) {
	tran.ReqInfo.ChannelId = channelId
}

/*设置 分页信息*/
func (tran *TranServices) SetPage(pageIndex, pageSize string) {
	tran.ReqInfo.PageIndex = pageIndex
	tran.ReqInfo.PageSize = pageSize
}

/*设置 对账单日期*/
func (tran *TranServices) SetBillDate(date string) {
	tran.ReqInfo.BillDate = date
}

/*设置 对账单类型*/
func (tran *TranServices) SetBillType(tp string) {
	tran.ReqInfo.BillType = tp
}

/*设置 附件数据*/
func (tran *TranServices) SetAttach(attach string) {
	tran.ReqInfo.Attach = attach
}

func (tran *TranServices) packMsg() ([]byte, error) {
	buf, err := xml.MarshalIndent(tran.ReqInfo, "", "    ")
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (tran *TranServices) unpackMsg(respBody []byte) error {
	err := xml.Unmarshal(respBody, tran.RspInfo)
	if err != nil {
		return err
	}
	return nil
}

func (tran *TranServices) clone() *TranServices {
	newTranSvr := TranServices{}

	newTranSvr = *tran
	newTranSvr.ReqInfo = new(Request)
	newTranSvr.RspInfo = new(Response)
	*newTranSvr.ReqInfo = *tran.ReqInfo
	*newTranSvr.RspInfo = *tran.RspInfo
	newTranSvr.WepayCfg = new(scanCfg.WepayCfg)
	*newTranSvr.WepayCfg = *tran.WepayCfg

	return &newTranSvr
}

func (tran *TranServices) InitTran(tran_cd, orig_tran_cd string, req *scanModel.TransMessage) gerror.IError {

	var err error

	//公共初始化
	tran.reqMsg = req
	tran.TranCode = tran_cd
	tran.ReqInfo.InterTranCode = tran.TranCode
	if req != nil && req.MsgBody != nil && req.MsgBody.Qr_code_info != nil &&
		req.MsgBody.Qr_code_info.Scance == "SYSTEM" {
		tran.SystemFlag = true
	}

	switch tran_cd + orig_tran_cd {
	case "1131": /*反扫*/
		tran.SetOrderId(req.MsgBody.Order_id)
		tran.SetTxnAmt(req.MsgBody.Tran_amt)
		tran.SetMerId(req.MsgBody.Mcht_cd)
		tran.SetBody(req.MsgBody.Mcht_nm)
		tran.SetAttach(req.MsgBody.Qr_code_info.Store_id)
		tran.SetGoodsTag(req.MsgBody.Qr_code_info.Goods_tag)
		tran.URL = tran.RemoteUrl + "pay/micropay"
		tran.ReqInfo.SpBillCreateIP = tran.OutIp
		tran.SetAuthCode(req.MsgBody.Qr_code_info.Auth_code)
	case "2131":
		fallthrough
	case "3131": /*扫码退款*/
		fallthrough
	case "3141":
		tran.SetTxnAmt(req.MsgBody.Tran_amt)
		tran.SetMerId(req.MsgBody.Mcht_cd)
		tran.SetBody(req.MsgBody.Mcht_nm)
		tran.SetGoodsTag(req.MsgBody.Qr_code_info.Goods_tag)
		tran.SetWxOrderId(req.MsgBody.Orig_sys_order_id)
		tran.SetOrderId(req.MsgBody.Orig_order_id)
		tran.SetRefundId(req.MsgBody.Order_id)
		tran.SetRefundAmt(req.MsgBody.Tran_amt)
		tran.SetOpUsrId(req.MsgBody.Mcht_cd)
		tran.URL = tran.RemoteUrl + "secapi/pay/refund"
	case "41311131": /*扫码冲正*/
		tran.SetTxnAmt(req.MsgBody.Tran_amt)
		tran.SetMerId(req.MsgBody.Mcht_cd)
		//tran.SetBody(req.MsgBody.Mcht_nm)
		tran.SetAttach(req.MsgBody.Qr_code_info.Store_id)
		tran.SetGoodsTag(req.MsgBody.Qr_code_info.Goods_tag)
		tran.SetOrderId(req.MsgBody.Orig_order_id)
		tran.URL = tran.RemoteUrl + "secapi/pay/reverse"
	case "41317131":
		fallthrough
	case "41311191":
		tran.SetMerId(req.MsgBody.Mcht_cd)
		tran.SetBody(req.MsgBody.Mcht_nm)
		tran.SetAttach(req.MsgBody.Qr_code_info.Store_id)
		tran.SetGoodsTag(req.MsgBody.Qr_code_info.Goods_tag)
		tran.SetOrderId(req.MsgBody.Orig_order_id)
		tran.URL = tran.RemoteUrl + "pay/closeorder"
	case "5131": /*订单查询*/
		//tran.SetTxnAmt(req.MsgBody.Tran_amt)
		tran.SetMerId(req.MsgBody.Mcht_cd)
		//tran.SetBody(req.MsgBody.Mcht_nm)
		tran.SetAttach(req.MsgBody.Qr_code_info.Store_id)
		tran.SetGoodsTag(req.MsgBody.Qr_code_info.Goods_tag)
		tran.URL = tran.RemoteUrl + "pay/orderquery"
		tran.SetOrderId(req.MsgBody.Order_id)
	case "5101": /*退款查询*/
		tran.URL = tran.RemoteUrl + "pay/refundquery"
	case "7131": /*传统正扫 创建订单*/
		tran.SetOrderId(req.MsgBody.Order_id)
		tran.SetTxnAmt(req.MsgBody.Tran_amt)
		tran.SetMerId(req.MsgBody.Mcht_cd)
		tran.SetBody(req.MsgBody.Mcht_nm)
		tran.SetAttach(req.MsgBody.Qr_code_info.Store_id)
		tran.SetGoodsTag(req.MsgBody.Qr_code_info.Goods_tag)
		tran.URL = tran.RemoteUrl + "pay/unifiedorder"
		tran.ReqInfo.TradeType = "NATIVE"
		tran.ReqInfo.SpBillCreateIP = tran.OutIp
		tran.ReqInfo.NotifyURL = tran.NotifyUrl
	case "1191": /*公众号支付 统一创建订单*/
		tran.SetOrderId(req.MsgBody.Order_id)
		tran.SetTxnAmt(req.MsgBody.Tran_amt)
		tran.SetMerId(req.MsgBody.Mcht_cd)
		tran.SetBody(req.MsgBody.Mcht_nm)
		tran.SetAttach(req.MsgBody.Qr_code_info.Store_id)
		tran.SetGoodsTag(req.MsgBody.Qr_code_info.Goods_tag)
		tran.URL = tran.RemoteUrl + "pay/unifiedorder"
		tran.ReqInfo.TradeType = "JSAPI"
		tran.ReqInfo.SpBillCreateIP = tran.OutIp
		tran.ReqInfo.NotifyURL = tran.NotifyUrl
		if len(req.MsgBody.Qr_code_info.Sub_app_id) > 0 {
			tran.SetSubAPPId(req.MsgBody.Qr_code_info.Sub_app_id)
			tran.SetSubOpenId(req.MsgBody.Qr_code_info.Sub_user_id)
		} else {
			tran.SetOpenId(req.MsgBody.Qr_code_info.User_id)
		}
	case "9001": /*商户入驻*/
		tran.URL = tran.RemoteUrl + "secapi/mch/submchmanage?action=add"
	case "9002": /*商户变更*/
		tran.URL = tran.RemoteUrl + "secapi/mch/submchmanage?action=modify"
	case "9003": /*下属商户查询*/
		tran.URL = tran.RemoteUrl + "secapi/mch/submchmanage?action=query"
	case "9004": /*子商户关联渠道商*/
		tran.URL = tran.RemoteUrl + "secapi/mch/channelsetting"
	case "9005": /*对账单下载*/
		tran.URL = tran.RemoteUrl + "pay/downloadbill"
	case "9006": /*查询子商户开发配置参数*/
		tran.URL = tran.RemoteUrl + "secapi/mch/querysubdevconfig"
	case "9007": /*新增创建子商户开发配置*/
		tran.URL = tran.RemoteUrl + "secapi/mch/addsubdevconfig"
	case "6131": /*订单通知*/
	default:
		return gerror.NewR(1002, nil, "非法交易码[%s]", tran_cd)
	}
	if err != nil {
		return gerror.NewR(1005, err, "交易处理失败")
	}

	return nil
}

func (tran *TranServices) DoTran(tran_cd, orig_tran_cd string) gerror.IError {

	var err error
	switch tran_cd + orig_tran_cd {
	case "1131": /*反扫*/
		if tran.SystemFlag {
			err = tran.FSServiceNoQuery()
		} else {
			err = tran.FSServiceWithQuery()
		}
	case "2131":
		fallthrough
	case "3131": /*扫码退款*/
		fallthrough
	case "3141":
		err = tran.RefundService()
	case "41311131": /*扫码冲正*/
		err = tran.CancelService()
	case "41317131":
		fallthrough
	case "41311191":
		err = tran.CloseService()
	case "5131": /*订单查询*/
		err = tran.QueryService()
	case "5101": /*退款查询*/
	case "7131": /*传统正扫 创建订单*/
		err = tran.ZSService()
	case "1191": /*公众号支付 统一创建订单*/
		err = tran.GZHService()
	case "9001": /*商户入驻*/
		err = tran.MchApplyService()
	case "9002": /*商户变更*/
		err = tran.MchModifyService()
	case "9003": /*下属商户查询*/
		err = tran.SubMchtSelService()
	case "9004":
		err = tran.DoWeComServices()
	case "9005": /*对账单下载*/
		err = tran.BillDownloadService()
	case "9006":
		fallthrough
	case "9007":
		err = tran.DoWeComServices()
	case "6131": /*订单通知*/
	default:
		return gerror.NewR(1002, nil, "非法交易码[%s]", tran_cd+orig_tran_cd)
	}
	if err != nil {
		return gerror.NewR(1005, err, "交易处理失败")
	}

	return nil
}

func (tran *TranServices) DoServices(req *scanModel.TransMessage) (*scanModel.TransMessage, gerror.IError) {

	var gerr gerror.IError
	chkType := "PAY"

	tran.Info("收到请求报文", req.Msg_body)

	switch req.MsgBody.Tran_cd {
	case "4131":
		gerr = tran.InitTran(req.MsgBody.Tran_cd, req.MsgBody.Orig_tran_cd, req)
	default:
		gerr = tran.InitTran(req.MsgBody.Tran_cd, "", req)
	}
	if gerr != nil {
		return nil, gerr
	}

	switch req.MsgBody.Tran_cd {
	case "4131":
		gerr = tran.DoTran(req.MsgBody.Tran_cd, req.MsgBody.Orig_tran_cd)
	case "2131":
		fallthrough
	case "3131":
		chkType = "REFUND"
		fallthrough
	default:
		gerr = tran.DoTran(req.MsgBody.Tran_cd, "")
	}
	if gerr != nil {
		return nil, gerr
	}

	//组应答报文
	rsp := req
	rsp.MsgBody.Resp_cd = tran.RespCd
	rsp.MsgBody.Resp_msg = tran.RespMsg
	rsp.MsgBody.ChnInsIdCd = tran.RspInfo.TransactionId
	rsp.MsgBody.Qr_code_info.Pay_bank = tran.RspInfo.BankType
	if len(tran.RspInfo.SubOpenId) > 0 {
		rsp.MsgBody.Qr_code_info.Open_id = tran.RspInfo.SubOpenId
	} else {
		rsp.MsgBody.Qr_code_info.Open_id = tran.RspInfo.OpenId
	}
	rsp.MsgBody.Qr_code_info.Pay_time = tran.RspInfo.TimeEnd
	rsp.MsgBody.Qr_code_info.Cash_amt = tran.RspInfo.CashFee
	rsp.MsgBody.Qr_code_info.Coupon_amt = tran.RspInfo.CouponFee
	rsp.MsgBody.Qr_code_info.Wx_jsapi = tran.WxJsStr
	rsp.MsgBody.Qr_code_info.Qr_code = tran.RspInfo.CodeURL

	//orig
	rsp.MsgBody.Orig_resp_cd = tran.OrigRespCd
	rsp.MsgBody.Orig_resp_msg = tran.OrigRespMsg
	rsp.MsgBody.Orig_sys_order_id = tran.RspInfo.OutTradeNo

	if rsp.MsgBody.Resp_cd == TRAN_SUCCESS {
		rsp.MsgBody.Ma_chk_key = req.MsgBody.InsIdCd + chkType + tran.ReqInfo.OutTradeNo
	}

	return rsp, nil
}

func (tran *TranServices) DoNotify(req []byte) ([]byte, gerror.IError) {

	tran.unpackMsg(req)
	tran.setRespInfo()
	tran.Debug(tran.RspToString())

	rsp := scanModel.TransMessage{}
	rsp.MsgBody = &scanModel.TransParams{}
	rsp.MsgBody.Tran_cd = "6131"
	rsp.MsgBody.InsIdCd = tran.ServerId
	rsp.MsgBody.Mcht_cd = tran.RspInfo.MchId
	rsp.MsgBody.Orig_resp_cd = tran.RespCd
	rsp.MsgBody.Orig_resp_msg = tran.RespMsg
	rsp.MsgBody.Sys_order_id = tran.RspInfo.TransactionId
	rsp.MsgBody.Order_id = tran.RspInfo.OutTradeNo
	rsp.MsgBody.Mcht_cd = tran.RspInfo.SubMchId
	rsp.MsgBody.Qr_code_info = &scanModel.QrCodeInfo{}
	rsp.MsgBody.Qr_code_info.Pay_bank = tran.RspInfo.BankType
	if len(tran.RspInfo.SubOpenId) > 0 {
		rsp.MsgBody.Qr_code_info.Open_id = tran.RspInfo.SubOpenId
	} else {
		rsp.MsgBody.Qr_code_info.Open_id = tran.RspInfo.OpenId
	}
	rsp.MsgBody.Qr_code_info.Pay_time = tran.RspInfo.TimeEnd
	rsp.MsgBody.Tran_amt = tran.RspInfo.TotalFee
	rsp.MsgBody.Qr_code_info.Cash_amt = tran.RspInfo.CashFee
	rsp.MsgBody.Qr_code_info.Coupon_amt = tran.RspInfo.CouponFee
	rsp.MsgBody.Ma_chk_key = tran.ServerId + "PAY" + tran.RspInfo.OutTradeNo

	//发送通知到核心
	go tran.SendNoti(&rsp)

	response := "<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>"
	return []byte(response), nil
}

func (tran *TranServices) DoBusSvr(req []byte) ([]byte, gerror.IError) {
	var gerr gerror.IError

	qu, err := url.ParseQuery(string(req))
	if err != nil {
		return nil, gerror.NewR(15001, err, "解析报文失败", string(req))
	}
	tran.Debug(qu)

	tranCd := qu.Get("Tran_cd")
	switch tranCd {
	case "9001":
		tran.SetMchName(qu.Get("MchntNm"))
		tran.SetMchShtName(qu.Get("MchntShortNm"))
		tran.SetServicePhone(qu.Get("MchntPhone"))
		tran.SetBussiness(qu.Get("Bussiness"))
		tran.SetMchRemark(qu.Get("MchntCd"))
		if qu.Get("ChannelId") == "" {
			tran.SetChannelId(tran.DefChnId)
		} else {
			tran.SetChannelId(qu.Get("ChannelId"))
		}
	case "9002":
		tran.SetMchShtName(qu.Get("MchntShortNm"))
		tran.SetMerId(qu.Get("MchntId"))
		tran.SetServicePhone(qu.Get("MchntPhone"))
	case "9003":
		tran.SetMchName(qu.Get("MchntNm"))
		tran.SetMerId(qu.Get("MchntId"))
		tran.SetPage(qu.Get("PageIndex"), qu.Get("PageSize"))
	case "9004":
		tran.SetMerId(qu.Get("MchntId"))
		if qu.Get("ChannelId") == "" {
			tran.SetChannelId(tran.DefChnId)
		} else {
			tran.SetChannelId(qu.Get("ChannelId"))
		}
	case "9005":
		tran.SetBillDate(qu.Get("BillDate"))
		tran.SetBillType(qu.Get("BillType"))
	case "9006":
		tran.SetMerId(qu.Get("MchntId"))
	case "9007":
		tran.SetMerId(qu.Get("MchntId"))
		tran.ReqInfo.SubAPPId = qu.Get("SubAppId")
		tran.ReqInfo.JsapiPath = qu.Get("JsapiPath")
		tran.ReqInfo.SubscribeAppid = qu.Get("SubscribeAppid")
	default:
		return nil, gerror.NewR(15010, nil, "不支持的类型", tranCd)
	}
	gerr = tran.InitTran(tranCd, "", nil)
	if gerr != nil {
		return nil, gerr
	}
	gerr = tran.DoTran(tranCd, "")
	if gerr != nil {
		return nil, gerr
	}

	msgRes := make(map[string]string, 0)
	switch tranCd {
	case "9001":
		fallthrough
	case "9002":
		msgRes["DstMchntCd"] = tran.RspInfo.SubMchId
		msgRes["RetMsg"] = tran.RespMsg + tran.RspInfo.ResultMsg
		msgRes["RetCd"] = tran.RespCd
	case "9003":
		msgRes["RetMsg"] = tran.RespMsg
		msgRes["RetCd"] = tran.RespCd
		if tran.RspInfo.SubMchInfs != nil {
			res, err := json.Marshal(tran.RspInfo.SubMchInfs)
			if err != nil {
				logr.Error("取子商户信息失败", err, tran.RspInfo.SubMchInfs)
			} else {
				msgRes["MchtInfos"] = string(res)
			}
		}
	case "9004":
		msgRes["RetMsg"] = tran.RespMsg
		msgRes["RetCd"] = tran.RespCd
		msgRes["SubMchId"] = tran.RspInfo.SubMchId
		msgRes["ChannelId"] = tran.RspInfo.ChanMchId
	case "9005":
		msgRes["RetCd"] = "00"
	case "9006":
		msgRes["RetMsg"] = tran.RespMsg
		msgRes["RetCd"] = tran.RespCd
		msgRes["AppConfList"] = tran.RspInfo.AppidConfList
		msgRes["JsapiPathList"] = tran.RspInfo.JsapiPathList
	case "9007":
		msgRes["RetMsg"] = tran.RespMsg
		msgRes["RetCd"] = tran.RespCd
	}

	//判断商户申请渠道号
	if tranCd == "9001" && msgRes["RetCd"] == "00" {
		gerr = tran.doChnCfg()
		if gerr != nil {
			msgRes["RetCd"] = gerr.GetErrorCode()
			msgRes["RetMsg"] = gerr.GetErrorString()
		}
	}

	res, err := json.Marshal(msgRes)
	if err != nil {
		return nil, gerror.NewR(17001, err, "生成应答失败")
	}

	return res, nil
}

func (tran *TranServices) ReqToString() string {
	pt := reflect.TypeOf(tran.ReqInfo).Elem()
	pv := reflect.ValueOf(tran.ReqInfo).Elem()
	var buf bytes.Buffer
	buf.WriteString("\n请求报文结构体开始:\n")
	for i := 0; i < pt.NumField(); i++ {
		pf := pv.Field(i)
		str := fmt.Sprintf("域名[%20s] 域值[%v];\n", pt.Field(i).Name, pf.Interface())
		buf.WriteString(str)
	}
	buf.WriteString("请求报文体结束;\n")
	return buf.String()
}

func (tran *TranServices) RspToString() string {
	pt := reflect.TypeOf(tran.RspInfo).Elem()
	pv := reflect.ValueOf(tran.RspInfo).Elem()
	var buf bytes.Buffer
	buf.WriteString("\n响应报文结构体开始:\n")
	for i := 0; i < pt.NumField(); i++ {
		pf := pv.Field(i)
		str := fmt.Sprintf("域名[%20s] 域值[%v];\n", pt.Field(i).Name, pf.Interface())
		buf.WriteString(str)
	}
	buf.WriteString("响应报文体结束;\n")
	return buf.String()
}

func (tran *TranServices) Comm(request []byte) ([]byte, error) {

	localAddr, err := net.ResolveTCPAddr("tcp4", tran.BindAddr+":0")
	if err != nil {
		tran.Errorf("tran.BindAddr[%s] 配置错误[%s];", tran.BindAddr, err)
		return nil, gerror.NewR(1004, err, "本地绑定地址配置错误;")
	}

	/*通信发送包*/
	tr := &http.Transport{
		Dial: (&net.Dialer{
			LocalAddr: localAddr,
			Timeout:   30 * time.Second,
			//KeepAlive: 30 * time.Second,
		}).Dial,
		TLSClientConfig: &tls.Config{Certificates: []tls.Certificate{tran.PrivateCert},
			InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: time.Second * time.Duration(tran.ServerTimeOut)}

	body := bytes.NewBuffer(request)
	req, _ := http.NewRequest("POST", tran.URL, body)
	req.Header.Set("Content-Type", "application/xml")

	out, _ := httputil.DumpRequestOut(req, true)
	tran.Debugf("http 请求包:\n-------------------\n [%s]\n-------------------\n ", string(out))

	resp, err := client.Do(req)
	if err != nil {
		tran.Errorf("POST: Client.Do error:[%s]", err)
		return nil, gerror.NewR(20020, err, "POST:client.Do error; ")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tran.Errorf("http.Status: %s is not success!", resp.Status)
		return nil, gerror.NewR(20030, nil, "POST:client.Do Status[%s] is not success; ", resp.StatusCode)
	}
	//读取应答
	data, _ := ioutil.ReadAll(resp.Body)
	tran.Debugf("响应报文：\n-------------------\n[%s]\n-------------------\n", string(data))

	return data, nil
}

/*微信交易应答码转换*/
func (tran *TranServices) setRespInfo() {

	/*首先判断通信应答码是否成功*/
	if tran.RspInfo.ReturnCode != COM_SUCCESS {
		tran.RespCd = LOCAl_COM_FAIL
		tran.RespMsg = tran.RspInfo.ReturnMsg
		return
	}

	/* 判断交易应答码是否成功*/
	if tran.TranCode == "5131" { /*查询交易*/
		/*转换应答码*/
		if tran.RspInfo.ResultCode != TRAN_SUCCESS {
			/*查询交易失败*/
			if len(tran.RspInfo.ErrCode) > 0 {
				/*错误码有存在*/
				errCd := tran.RspInfo.ErrCode
				if respInfo, ok := RespCdMap[errCd]; ok {
					tran.Infof("订单号[%s]转换应答码[%s]到[%s][%s];",
						tran.ReqInfo.OutTradeNo, errCd, respInfo.Code, respInfo.Desc)
					tran.RespCd = respInfo.Code
					tran.RespMsg = tran.RspInfo.ErrCodeDes
				} else {
					tran.RespCd = OTHER_BUSI_FAIL
					tran.RespMsg = tran.RspInfo.ErrCodeDes
					tran.Infof("订单号[%s]没有匹配到应答码使用[%s];", tran.ReqInfo.OutTradeNo, OTHER_BUSI_FAIL)
				}
				return
			} else {
				tran.RespCd = OTHER_BUSI_FAIL
				tran.RespMsg = tran.RspInfo.ErrCodeDes
				tran.Infof("订单号[%s]没有返回ErrCode, 使用[%s];", tran.ReqInfo.OutTradeNo, OTHER_BUSI_FAIL)
				return
			}
		} else {
			/*查询交易成功*/
			tran.RespCd = LOCAL_BUSI_SUCCESS
			tran.RespMsg = TRAN_SUCCESS
			/*转换交易状态*/
			tradeStatus := tran.RspInfo.TradeState
			if tradeStatus == TRAN_SUCCESS {
				tran.OrigRespCd = LOCAL_BUSI_SUCCESS
				tran.OrigRespMsg = TRAN_SUCCESS
			} else {
				if respInfo, ok := RespCdMap[tradeStatus]; ok {
					tran.Infof("订单号[%s]交易状态转换[%s]到应答码[%s][%s];",
						tran.ReqInfo.OutTradeNo, tradeStatus, respInfo.Code, respInfo.Desc)
					tran.OrigRespCd = respInfo.Code
					tran.OrigRespMsg = tran.RspInfo.TradeStateDes
				} else {
					tran.OrigRespCd = OTHER_BUSI_FAIL
					tran.OrigRespMsg = tran.RspInfo.TradeStateDes
					tran.Infof("订单号[%s]交易状态转换[%s]没有匹配到应答码[%s];", tran.ReqInfo.OutTradeNo, tradeStatus, OTHER_BUSI_FAIL)
				}
			}
			return
		}
	} else { /*非查询类交易*/
		if tran.RspInfo.ResultCode != TRAN_SUCCESS {
			/*失败交易*/
			if len(tran.RspInfo.ErrCode) > 0 {
				/*错误码有存在*/
				errCd := tran.RspInfo.ErrCode
				if respInfo, ok := RespCdMap[errCd]; ok {
					tran.Infof("订单号[%s]转换应答码[%s]到[%s][%s];",
						tran.ReqInfo.OutTradeNo, errCd, respInfo.Code, respInfo.Desc)
					tran.RespCd = respInfo.Code
					tran.RespMsg = tran.RspInfo.ErrCodeDes
				} else {
					tran.RespCd = OTHER_BUSI_FAIL
					tran.RespMsg = tran.RspInfo.ErrCodeDes
					tran.Infof("订单号[%s]没有匹配到应答码使用[%s];", tran.ReqInfo.OutTradeNo, OTHER_BUSI_FAIL)
				}
				return
			} else {
				/*没有错误码存在*/
				tran.RespCd = OTHER_BUSI_FAIL
				tran.RespMsg = tran.RspInfo.ErrCodeDes
				tran.Infof("订单号[%s]没有返回ErrCode, 使用[%s];", tran.ReqInfo.OutTradeNo, OTHER_BUSI_FAIL)
				return
			}
		} else {
			/*成功交易*/
			tran.RespCd = LOCAL_BUSI_SUCCESS
			tran.RespMsg = TRAN_SUCCESS
			return
		}
	}
	return
}

func (tran *TranServices) setWxJsapi() error {
	jsApi := new(WxJsapi)
	jsApi.APPID = tran.AppId
	jsApi.TimeStamp = fmt.Sprintf("%d", time.Now().Unix())

	jsApi.Package = fmt.Sprintf("prepay_id=%s", tran.RspInfo.PrePayId)
	jsApi.SignType = "MD5"

	err := tran.SignJsapi(jsApi)
	if err != nil {
		tran.Errorf("SingleServer.SignJsapi 签名失败[%s];", err)
		return gerror.NewR(10050, err, "SingleServer.SignJsapi 签名失败")
	}

	buf, err := json.MarshalIndent(jsApi, "", "")
	if err != nil {
		tran.Errorf("json.MarshalIndent 生成失败[%s];", err)
		return gerror.NewR(10050, err, "json.MarshalIndent 生成失败")
	}
	tran.WxJsStr = string(buf)
	tran.Debug("公众号返回JSAPI串:[%s]", tran.WxJsStr)
	return nil
}

/*通用调用微信服务*/
func (tran *TranServices) callWxServer() error {
	var err error

	//签名
	err = tran.Sign()
	if err != nil {
		tran.Errorf("签名失败[%s]；", err)
		return gerror.NewR(1006, err, "签名失败[%s]", err)
	}
	//打包
	sndBuf, err := tran.packMsg()
	if err != nil {
		tran.Errorf("打包失败[%s]；", err)
		return gerror.NewR(1006, err, "打包失败[%s]", err)
	}
	//打印请求报文
	tran.Debug(tran.ReqToString())
	/*发送报文*/
	respBody, err := tran.Comm(sndBuf)
	if err != nil {
		tran.Errorf("和对端通信失败[%s]；", err)
		return gerror.NewR(1006, err, "通信失败")
	}
	//解包报文
	err = tran.unpackMsg(respBody)
	if err != nil {
		tran.Errorf("解包失败:[%s]；", err)
		return gerror.NewR(1006, err, "解包失败")
	}

	/*更新内部应答码*/
	TrnCd, _ := strconv.ParseInt(tran.ReqInfo.InterTranCode, 10, 0)
	tran.RspInfo.InterTranCode = fmt.Sprintf("%04d", TrnCd+1)
	//打印响应报文
	tran.Debug(tran.RspToString())

	/*
		//验签
		err = SingleServer.Verify(tran)
		if err != nil {
			tran.Error("订单号[%s]验签失败[%s];", tran.ReqInfo.OutTradeNo, err)
			return gerror.NewR(1006, err, "订单号[%s]验签失败", tran.ReqInfo.OutTradeNo)
		}
	*/
	return nil

}

/*反扫交易处理 不再自主查询
由前端机构发起查询*/
func (tran *TranServices) FSServiceNoQuery() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "1131" {
		return gerror.NewR(1003, nil, "订单号[%s] 非法交易吗[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}

	/*调用服务*/
	err = tran.callWxServer()
	if err != nil {
		tran.Errorf("FSServiceNoQuery: [%s]callWxServer 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1003, err, "FSServiceNoQuery: [%s]callWxServer 调用失败;", tran.ReqInfo.OutTradeNo)
	}
	/*响应报文处理*/
	tran.setRespInfo()

	return nil
}

/*反扫交易处理 由自主发起查询
由前端机构发起查询*/
func (tran *TranServices) FSServiceWithQuery() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "1131" {
		return gerror.NewR(1004, nil, "订单号[%s] 非法交易吗[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}
	/*调用服务*/
	err = tran.callWxServer()
	if err != nil {
		tran.Errorf("FSServiceWithQuery: [%s]callWxServer 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1004, err, "FSServiceNoQuery: [%s]callWxServer 调用失败;", tran.ReqInfo.OutTradeNo)
	}
	/*响应报文处理*/
	tran.setRespInfo()

	/*需要发起查询*/
	if tran.RespCd == "W3" {
		i := 0
		cancelFlag := true
		for ; i < tran.QueryCnt; i++ {
			time.Sleep(time.Second * time.Duration(tran.QueryInt))
			tran.Infof("FSServiceWithQuery: 订单号[%s] 返回结果[%s]; 开发发起第[%d]次查询;",
				tran.ReqInfo.OutTradeNo, tran.RespCd, i+1)
			query, gerr := NewTranServices(tran.WepayCfg)
			if gerr != nil {
				tran.RespCd = "98"
				tran.RespMsg = "交易状态未知，请稍后查询"
				tran.Warnf("FSServiceWithQuery: 订单号[%s] 查询失败[%v] [%s][%s]", tran.ReqInfo.OutTradeNo, gerr,
					tran.RespCd, tran.RespMsg)
				return nil
			}
			query.InitTran("5131", "", tran.reqMsg)
			query.SetMerId(tran.ReqInfo.SubMchId)
			query.SetOrderId(tran.ReqInfo.OutTradeNo)
			err := query.DoTran("5131", "")
			if err != nil {
				tran.Errorf("FSServiceNoQuery: [%s]QueryService 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
				continue
			}
			/*响应报文处理*/
			query.setRespInfo()
			tran.Infof("FSServiceWithQuery: 订单号[%s]第[%d]次查询[%s][%s]原交易[%s][%s];", tran.ReqInfo.OutTradeNo, i+1,
				query.RespCd, query.RespMsg,
				query.OrigRespCd, query.OrigRespMsg)
			/*查询交易不成功*/
			if query.RespCd != LOCAL_BUSI_SUCCESS {
				tran.Infof("FSServiceWithQuery: 订单号[%s]第[%d]次查询[%s][%s] 结果不成功，继续查询;", tran.ReqInfo.OutTradeNo, i+1,
					query.RespCd, query.RespMsg)
				continue
			}

			if query.OrigRespCd != "W3" && len(query.OrigRespCd) > 0 {
				tran.RespCd = query.OrigRespCd
				tran.RespMsg = query.OrigRespMsg
				tran.RspInfo.TransactionId = query.RspInfo.TransactionId
				tran.RspInfo.OutTradeNo = query.RspInfo.OutTradeNo
				tran.RspInfo.BankType = query.RspInfo.BankType
				tran.RspInfo.OpenId = query.RspInfo.OpenId
				tran.RspInfo.SubAPPId = query.RspInfo.SubAPPId
				tran.RspInfo.SubOpenId = query.RspInfo.SubOpenId
				tran.RspInfo.TimeEnd = query.RspInfo.TimeEnd

				/*不再撤销*/
				cancelFlag = false
				break
			}
		}
		if i == tran.QueryCnt && cancelFlag {
			//		tran.Infof("FSServiceNoQuery: 开始调用撤销交易，撤销订单[%s];", tran.ReqInfo.OutTradeNo)
			//		cancelOk := false
			//		for i := 0; i < tran.QueryCnt; i++ {
			//			time.Sleep(time.Second * time.Duration(tran.QueryInt))
			//			cancel, gerr := NewTranServices(tran.WepayCfg)
			//			if gerr != nil {
			//				tran.RespCd = "98"
			//				tran.RespMsg = "交易状态未知，请稍后查询"
			//				tran.Warnf("FSServiceWithQuery: 订单号[%s] 撤销失败[%v] [%s][%s]", tran.ReqInfo.OutTradeNo, gerr,
			//					tran.RespCd, tran.RespMsg)
			//				return nil
			//			}
			//			cancel.InitTran("4131", "1131", tran.reqMsg)
			//			cancel.SetMerId(tran.ReqInfo.SubMchId)
			//			cancel.SetOrderId(tran.ReqInfo.OutTradeNo)
			//			err := cancel.CancelService()
			//			if err != nil {
			//				tran.Errorf("FSServiceNoQuery: [%s] CancelService 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
			//				continue
			//			}
			//			tran.Infof("FSServiceWithQuery: 订单号[%s]第[%d]次撤销[%s][%s];", tran.ReqInfo.OutTradeNo, i+1,
			//				cancel.RespCd, cancel.RespMsg)
			//			if cancel.RspInfo.ReCall == "N" {
			//				cancelOk = true
			//				break
			//			}
			//		}
			//		if cancelOk {
			//			/*置原交易为已撤销*/
			//			tran.RespCd = "S5"
			//			tran.RespMsg = "订单已撤销"
			//			tran.Infof("FSServiceWithQuery: 订单号[%s] 已撤销返回前端应答码[%s][%s]", tran.ReqInfo.OutTradeNo,
			//				tran.RespCd, tran.RespMsg)
			//		} else {
			tran.RespCd = "98"
			tran.RespMsg = "交易状态未知，请稍后查询"
			//tran.Warnf("FSServiceWithQuery: 订单号[%s] 撤销失败 [%s][%s]", tran.ReqInfo.OutTradeNo,
			//				tran.RespCd, tran.RespMsg)
			//		}
			return nil
		}
	}

	return nil
}

/*查询交易*/
func (tran *TranServices) QueryService() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "5131" {
		return gerror.NewR(1005, nil, "订单号[%s] 非法交易吗[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}
	err = tran.callWxServer()
	if err != nil {
		tran.Errorf("QueryService: [%s]callWxServer 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1005, err, "QueryService: [%s]callWxServer 调用失败;", tran.ReqInfo.OutTradeNo)
	}
	/*响应报文处理*/
	tran.setRespInfo()
	return nil
}

/*撤销交易-对应冲正*/
func (tran *TranServices) CancelService() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "4131" {
		return gerror.NewR(1006, nil, "订单号[%s] 非法交易吗[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}
	err = tran.callWxServer()
	if err != nil {
		tran.Errorf("CancelService: [%s]callWxServer 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1006, err, "CancelService: [%s]callWxServer 调用失败;", tran.ReqInfo.OutTradeNo)
	}
	tran.setRespInfo()
	return nil
}

/*退款交易处理*/
func (tran *TranServices) RefundService() error {
	var err error
	err = tran.callWxServer()
	if err != nil {
		tran.Errorf("RefundService: [%s]callWxServer 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1007, err, "RefundService: [%s]callWxServer 调用失败;", tran.ReqInfo.OutTradeNo)
	}
	tran.setRespInfo()
	return nil
}

/*撤销查询交易*/
func (tran *TranServices) RefundQueryService() error {
	var err error
	err = tran.callWxServer()
	if err != nil {
		tran.Errorf("RefundQueryService: [%s]callWxServer 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1008, err, "RefundQueryService: [%s]callWxServer 调用失败;", tran.ReqInfo.OutTradeNo)
	}
	tran.setRespInfo()
	return nil
}

/*正扫交易处理*/
func (tran *TranServices) ZSService() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "7131" {
		return gerror.NewR(1009, nil, "订单号[%s] 非法交易吗[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}
	err = tran.callWxServer()
	if err != nil {
		tran.Errorf("ZSService: [%s]callWxServer 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1009, err, "ZSService: [%s]callWxServer 调用失败;", tran.ReqInfo.OutTradeNo)
	}
	/*应答码转换*/
	tran.setRespInfo()

	return nil
}

/*关闭订单交易处理*/
func (tran *TranServices) CloseService() error {
	var err error
	err = tran.callWxServer()
	if err != nil {
		tran.Errorf("CloseService: [%s]callWxServer 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1010, err, "CloseService: [%s]callWxServer 调用失败;", tran.ReqInfo.OutTradeNo)
	}
	tran.setRespInfo()
	return nil
}

/*公众号支付*/
func (tran *TranServices) GZHService() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "1191" {
		return gerror.NewR(1010, nil, "订单号[%s] 非法交易吗[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}

	err = tran.callWxServer()
	if err != nil {
		tran.Errorf("GZHService: [%s]callWxServer 调用失败:[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1011, err, "GZHService: [%s]callWxServer 调用失败;", tran.ReqInfo.OutTradeNo)
	}
	tran.setRespInfo()
	/*组装WxJSAPI串*/
	if tran.RespCd == LOCAL_BUSI_SUCCESS {
		err := tran.setWxJsapi()
		if err != nil {
			tran.Errorf("setWxJsapi 调用失败:[%s]", err)
			return gerror.NewR(1011, err, "setWxJsapi 调用失败")
		}
	}
	return nil
}

/*商户入驻*/
func (tran *TranServices) MchApplyService() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "9001" {
		return gerror.NewR(1009, nil, "订单号[%s] 非法交易码[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}

	//签名
	err = tran.SignMcht()
	if err != nil {
		tran.Errorf("签名失败[%s]；", err)
		return gerror.NewR(1006, err, "签名失败[%s]", err)
	}
	//打包
	sndBuf, err := tran.packMsg()
	if err != nil {
		tran.Errorf("打包失败[%s]；", err)
		return gerror.NewR(1006, err, "打包失败[%s]", err)
	}
	//打印请求报文
	tran.Debug(tran.ReqToString())
	/*发送报文*/
	respBody, err := tran.Comm(sndBuf)
	if err != nil {
		tran.Errorf("和对端通信失败[%s]；", err)
		return gerror.NewR(1006, err, "通信失败")
	}
	//解包报文
	err = tran.unpackMsg(respBody)
	if err != nil {
		tran.Errorf("解包失败:[%s]；", err)
		return gerror.NewR(1006, err, "解包失败")
	}

	/*更新内部应答码*/
	TrnCd, _ := strconv.ParseInt(tran.ReqInfo.InterTranCode, 10, 0)
	tran.RspInfo.InterTranCode = fmt.Sprintf("%04d", TrnCd+1)
	//打印响应报文
	tran.Debug(tran.RspToString())

	//验签
	err = tran.Verify()
	if err != nil {
		tran.Errorf("订单号[%s]验签失败[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1006, err, "订单号[%s]验签失败", tran.ReqInfo.OutTradeNo)
	}
	/*应答码转换*/
	tran.setRespInfo()

	return nil
}

/*商户变更*/
func (tran *TranServices) MchModifyService() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "9002" {
		return gerror.NewR(1009, nil, "订单号[%s] 非法交易码[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}

	//签名
	err = tran.SignMcht()
	if err != nil {
		tran.Errorf("签名失败[%s]；", err)
		return gerror.NewR(1006, err, "签名失败[%s]", err)
	}
	//打包
	sndBuf, err := tran.packMsg()
	if err != nil {
		tran.Errorf("打包失败[%s]；", err)
		return gerror.NewR(1006, err, "打包失败[%s]", err)
	}
	//打印请求报文
	tran.Debug(tran.ReqToString())
	/*发送报文*/
	respBody, err := tran.Comm(sndBuf)
	if err != nil {
		tran.Errorf("和对端通信失败[%s]；", err)
		return gerror.NewR(1006, err, "通信失败")
	}
	//解包报文
	err = tran.unpackMsg(respBody)
	if err != nil {
		tran.Errorf("解包失败:[%s]；", err)
		return gerror.NewR(1006, err, "解包失败")
	}

	/*更新内部应答码*/
	TrnCd, _ := strconv.ParseInt(tran.ReqInfo.InterTranCode, 10, 0)
	tran.RspInfo.InterTranCode = fmt.Sprintf("%04d", TrnCd+1)
	//打印响应报文
	tran.Debug(tran.RspToString())

	//验签
	err = tran.Verify()
	if err != nil {
		tran.Errorf("订单号[%s]验签失败[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1006, err, "订单号[%s]验签失败", tran.ReqInfo.OutTradeNo)
	}
	/*应答码转换*/
	tran.setRespInfo()

	return nil
}

/*下属商户查询*/
func (tran *TranServices) SubMchtSelService() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "9003" {
		return gerror.NewR(1009, nil, "订单号[%s] 非法交易码[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}

	//签名
	err = tran.SignMcht()
	if err != nil {
		tran.Errorf("签名失败[%s]；", err)
		return gerror.NewR(1006, err, "签名失败[%s]", err)
	}
	//打包
	sndBuf, err := tran.packMsg()
	if err != nil {
		tran.Errorf("打包失败[%s]；", err)
		return gerror.NewR(1006, err, "打包失败[%s]", err)
	}
	//打印请求报文
	tran.Debug(tran.ReqToString())
	/*发送报文*/
	respBody, err := tran.Comm(sndBuf)
	if err != nil {
		tran.Errorf("和对端通信失败[%s]；", err)
		return gerror.NewR(1006, err, "通信失败")
	}
	//解包报文
	err = tran.unpackMsg(respBody)
	if err != nil {
		tran.Errorf("解包失败:[%s]；", err)
		return gerror.NewR(1006, err, "解包失败")
	}

	/*更新内部应答码*/
	TrnCd, _ := strconv.ParseInt(tran.ReqInfo.InterTranCode, 10, 0)
	tran.RspInfo.InterTranCode = fmt.Sprintf("%04d", TrnCd+1)
	//打印响应报文
	tran.Debug(tran.RspToString())

	//验签
	err = tran.Verify()
	if err != nil {
		tran.Errorf("订单号[%s]验签失败[%s];", tran.ReqInfo.OutTradeNo, err)
		return gerror.NewR(1006, err, "订单号[%s]验签失败", tran.ReqInfo.OutTradeNo)
	}
	/*应答码转换*/
	tran.setRespInfo()

	return nil
}

/*对账单下载*/
func (tran *TranServices) BillDownloadService() error {
	var err error
	/*校验交易码*/
	if tran.TranCode != "9005" {
		return gerror.NewR(1009, nil, "订单号[%s] 非法交易码[%s]调用", tran.ReqInfo.OutTradeNo, tran.TranCode)
	}

	//签名
	err = tran.Sign()
	if err != nil {
		tran.Errorf("签名失败[%s]；", err)
		return gerror.NewR(1006, err, "签名失败[%s]", err)
	}
	//打包
	sndBuf, err := tran.packMsg()
	if err != nil {
		tran.Errorf("打包失败[%s]；", err)
		return gerror.NewR(1006, err, "打包失败[%s]", err)
	}
	//打印请求报文
	tran.Debug(tran.ReqToString())
	/*发送报文*/
	respBody, err := tran.Comm(sndBuf)
	if err != nil {
		tran.Errorf("和对端通信失败[%s]；", err)
		return gerror.NewR(1006, err, "通信失败")
	}
	//解包报文
	err = tran.unpackMsg(respBody)
	if err != nil {
		dir := tran.BillDir + "/" + tran.ReqInfo.BillDate
		localFile := dir + "/" + tran.ServerId + "_" + tran.ReqInfo.BillDate + ".csv"
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			tran.Errorf("目录[%s]创建失败[%s];", dir, err)
			return gerror.NewR(30030, err, "本地创建目录失败[%s]", dir)
		}
		ioutil.WriteFile(localFile, respBody, 0664)

		tran.Infof("[%s] 对账文件下载成功;", tran.ReqInfo.BillDate)
		return nil
	}
	//下载失败
	return gerror.NewR(1007, err, "下载对账文件失败")
}

func (tran *TranServices) SignJsapi(jsApi *WxJsapi) error {
	//生成随机串
	jsApi.NonceStr, _ = getRandom()
	//计算签名
	pt := reflect.TypeOf(jsApi).Elem()
	pv := reflect.ValueOf(jsApi).Elem()

	keyArr := make([]string, 20)
	valMap := make(map[string]string, 20)
	for i := 0; i < pt.NumField(); i++ {
		pf := pv.Field(i)
		tag := pt.Field(i).Tag.Get("json")
		xmlTag := parseTag(tag)
		switch pf.Kind() {
		case reflect.String:
			keyArr = append(keyArr, xmlTag)
			valMap[xmlTag] = pf.String()
		case reflect.Int:
			if pf.Int() > 0 {
				keyArr = append(keyArr, xmlTag)
				valMap[xmlTag] = fmt.Sprintf("%d", pf.Int())
			}
		default:
			break
		}
	}
	sort.Strings(keyArr)
	var signBuf bytes.Buffer
	for _, key := range keyArr {
		if len(valMap[key]) > 0 {
			s := fmt.Sprintf("%s=%s&", key, valMap[key])
			signBuf.WriteString(s)
		}
	}
	key := fmt.Sprintf("key=%s", tran.AppSecret)
	signBuf.WriteString(key)
	tran.Debug("JSAPI调用签名串[%s]", signBuf.String())
	jsApi.PaySign = fmt.Sprintf("%X", md5.Sum(signBuf.Bytes()))
	tran.Debug("JSAPI签名结果[%s]", jsApi.PaySign)
	return nil
}

func (tran *TranServices) Sign() error {
	//生成随机串
	var err error
	tran.ReqInfo.NonceStr, err = getRandom()
	if err != nil {
		tran.Errorf("getRandom error[%s]", err)
		return err
	}

	//计算签名
	pt := reflect.TypeOf(tran.ReqInfo).Elem()
	pv := reflect.ValueOf(tran.ReqInfo).Elem()

	keyArr := make([]string, 20)
	valMap := make(map[string]string, 20)
	for i := 0; i < pt.NumField(); i++ {
		pf := pv.Field(i)
		tag := pt.Field(i).Tag.Get("xml")
		xmlTag := parseTag(tag)
		/*不参与打包交易不参与签名*/
		if xmlTag == "-" {
			continue
		}
		switch pf.Kind() {
		case reflect.String:
			keyArr = append(keyArr, xmlTag)
			valMap[xmlTag] = pf.String()
		case reflect.Int:
			if pf.Int() > 0 {
				keyArr = append(keyArr, xmlTag)
				valMap[xmlTag] = fmt.Sprintf("%d", pf.Int())
			}
		default:
			break
		}
	}
	sort.Strings(keyArr)
	var signBuf bytes.Buffer
	for _, key := range keyArr {
		if len(valMap[key]) > 0 {
			s := fmt.Sprintf("%s=%s&", key, valMap[key])
			signBuf.WriteString(s)
		}
	}
	key := fmt.Sprintf("key=%s", tran.AppSecret)
	signBuf.WriteString(key)
	tran.Debugf("[%s]", signBuf.String())
	tran.ReqInfo.Sign = fmt.Sprintf("%X", md5.Sum(signBuf.Bytes()))
	return nil
}

func (tran *TranServices) SignMcht() error {

	//计算签名
	pt := reflect.TypeOf(tran.ReqInfo).Elem()
	pv := reflect.ValueOf(tran.ReqInfo).Elem()

	keyArr := make([]string, 20)
	valMap := make(map[string]string, 20)
	for i := 0; i < pt.NumField(); i++ {
		pf := pv.Field(i)
		tag := pt.Field(i).Tag.Get("xml")
		xmlTag := parseTag(tag)
		/*不参与打包交易不参与签名*/
		if xmlTag == "-" {
			continue
		}
		switch pf.Kind() {
		case reflect.String:
			keyArr = append(keyArr, xmlTag)
			valMap[xmlTag] = pf.String()
		case reflect.Int:
			if pf.Int() > 0 {
				keyArr = append(keyArr, xmlTag)
				valMap[xmlTag] = fmt.Sprintf("%d", pf.Int())
			}
		default:
			break
		}
	}
	sort.Strings(keyArr)
	var signBuf bytes.Buffer
	for _, key := range keyArr {
		if len(valMap[key]) > 0 {
			s := fmt.Sprintf("%s=%s&", key, valMap[key])
			signBuf.WriteString(s)
		}
	}
	key := fmt.Sprintf("key=%s", tran.AppSecret)
	signBuf.WriteString(key)
	tran.Debug("[%s]", signBuf.String())
	tran.ReqInfo.Sign = fmt.Sprintf("%X", md5.Sum(signBuf.Bytes()))
	return nil
}

func (tran *TranServices) Verify() error {
	if len(tran.RspInfo.Sign) == 0 {
		tran.Errorf("响应报文[%s]里没有签名,无需验签;", tran.ReqInfo.OutTradeNo)
		return nil
	}
	//计算签名
	pt := reflect.TypeOf(tran.RspInfo).Elem()
	pv := reflect.ValueOf(tran.RspInfo).Elem()

	keyArr := make([]string, 20)
	valMap := make(map[string]string, 20)
	for i := 0; i < pt.NumField(); i++ {
		pf := pv.Field(i)
		tag := pt.Field(i).Tag.Get("xml")
		xmlTag := parseTag(tag)
		/*不参与打包交易不参与签名*/
		if xmlTag == "-" {
			continue
		}

		if xmlTag == "sign" {
			continue
		}
		switch pf.Kind() {
		case reflect.String:
			keyArr = append(keyArr, xmlTag)
			valMap[xmlTag] = pf.String()
		case reflect.Int:
			if pf.Int() > 0 {
				keyArr = append(keyArr, xmlTag)
				valMap[xmlTag] = fmt.Sprintf("%d", pf.Int())
			}
		default:
			break
		}
	}
	sort.Strings(keyArr)
	var signBuf bytes.Buffer
	for _, key := range keyArr {
		if len(valMap[key]) > 0 {
			s := fmt.Sprintf("%s=%s&", key, valMap[key])
			signBuf.WriteString(s)
		}
	}
	key := fmt.Sprintf("key=%s", tran.AppSecret)
	signBuf.WriteString(key)
	tran.Debug("验签报文[%s]", signBuf.String())
	localeSign := fmt.Sprintf("%X", md5.Sum(signBuf.Bytes()))
	if localeSign != tran.RspInfo.Sign {
		tran.Errorf("验签失败[%s][%s]", localeSign, tran.RspInfo.Sign)
		return gerror.NewR(1005, nil, "验签失败")
	} else {
		tran.Debug("验签成功[%s][%s]", localeSign, tran.RspInfo.Sign)
	}

	return nil
}

func (tran *TranServices) SendNoti(msg *scanModel.TransMessage) gerror.IError {

	if tran.NotiUrl == "" {
		return nil
	}

	rsp, gerr := msg.PackReq()
	if gerr != nil {
		return gerr
	}
	for i := 0; i < tran.NotiNum; i++ {
		res, gerr := tran.DoComm(rsp, tran.NotiUrl)
		if gerr == nil {
			resMsg := scanModel.TransMessage{}
			err := json.Unmarshal(res, &resMsg)
			if err != nil {
				tran.Warn("解析应答失败", res)
			}
			err = json.Unmarshal([]byte(resMsg.Msg_body), &resMsg.MsgBody)
			if err != nil {
				tran.Warn("解析应答失败", res)
			}
			if resMsg.MsgBody.Resp_cd == "00" {
				return nil
			}
			tran.Debug("交易应答失败", resMsg.Msg_body)
		}
		time.Sleep(time.Duration((i+1)*5) * time.Second)
	}
	tran.Warn("异步通知失败达最大次数，不再发送", string(rsp))

	return nil
}

func (tran *TranServices) DoComm(request []byte, remoteAddr string) ([]byte, error) {

	tran.Debug("remoteurl: ", remoteAddr)
	/*通信发送包*/
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 30 * time.Second,
			//KeepAlive: 30 * time.Second,
		}).Dial,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: time.Second * time.Duration(tran.ServerTimeOut)}

	body := bytes.NewBuffer(request)
	req, err := http.NewRequest("POST", remoteAddr, body)
	if err != nil {
		return nil, gerror.NewR(1005, err, "请求失败", remoteAddr)
	}
	req.Header.Set("Content-Type", "application/json")

	tran.Debug("http 请求包:", string(request))

	resp, err := client.Do(req)
	if err != nil {
		tran.Errorf("POST: Client.Do error:[%s]", err)
		return nil, gerror.NewR(20020, err, "POST:client.Do error; ")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tran.Errorf("http.Status: %s is not success!", resp.Status)
		return nil, gerror.NewR(20030, nil, "POST:client.Do Status[%s] is not success; ", resp.StatusCode)
	}
	//读取应答
	data, _ := ioutil.ReadAll(resp.Body)
	tran.Debugf("响应报文：\n-------------------\n[%s]\n-------------------\n", string(data))

	return data, nil
}

func (tran *TranServices) DoWeComServices() error {
	var err error

	//签名
	err = tran.SignMcht()
	if err != nil {
		tran.Errorf("签名失败[%s]；", err)
		return gerror.NewR(1006, err, "签名失败[%s]", err)
	}
	//打包
	sndBuf, err := tran.packMsg()
	if err != nil {
		tran.Errorf("打包失败[%s]；", err)
		return gerror.NewR(1006, err, "打包失败[%s]", err)
	}
	//打印请求报文
	tran.Debug(tran.ReqToString())
	/*发送报文*/
	respBody, err := tran.Comm(sndBuf)
	if err != nil {
		tran.Errorf("和对端通信失败[%s]；", err)
		return gerror.NewR(1006, err, "通信失败")
	}
	//解包报文
	err = tran.unpackMsg(respBody)
	if err != nil {
		tran.Errorf("解包失败:[%s]；", err)
		return gerror.NewR(1006, err, "解包失败")
	}

	/*更新内部应答码*/
	//TrnCd, _ := strconv.ParseInt(tran.ReqInfo.InterTranCode, 10, 0)
	//tran.RspInfo.InterTranCode = fmt.Sprintf("%04d", TrnCd+1)
	//打印响应报文
	tran.Debug(tran.RspToString())

	//验签
	//err = tran.Verify()
	//if err != nil {
	//	tran.Errorf("订单号[%s]验签失败[%s];", tran.ReqInfo.OutTradeNo, err)
	//	return gerror.NewR(1006, err, "订单号[%s]验签失败", tran.ReqInfo.OutTradeNo)
	//}
	/*应答码转换*/
	tran.setRespInfo()

	return nil
}

func (tran *TranServices) doChnCfg() gerror.IError {

	if tran.WeChnCfg == nil {
		tran.Debug("无渠道配置，不处理")
		return nil
	}
	cfg, ok := tran.WeChnCfg[tran.ReqInfo.ChannelId]
	if !ok {
		return nil
	}

	var T9007 = fmt.Sprintf("Tran_cd=9007&MchntId=%s&", tran.RspInfo.SubMchId)

	if cfg.JsapiPath != "" {
		jsapiList := strings.Split(cfg.JsapiPath, ",")
		for _, js := range jsapiList {
			req := T9007 + fmt.Sprintf("JsapiPath=%s", url.QueryEscape(js))
			tsvr, gerr := NewTranServices(tran.WepayCfg)
			if gerr != nil {
				return gerr
			}
			_, gerr = tsvr.DoBusSvr([]byte(req))
			if gerr != nil {
				return gerr
			}
			if tsvr.RespCd != "00" && !strings.Contains(tsvr.RespMsg, "已存在") {
				return gerror.NewR(20040, nil, tsvr.RespMsg)
			}
		}
		//time.Sleep(time.Second)
	}

	if cfg.SubscribeAppid != "" {
		req := T9007 + fmt.Sprintf("SubscribeAppid=%s", cfg.SubscribeAppid)
		tsvr, gerr := NewTranServices(tran.WepayCfg)
		if gerr != nil {
			return gerr
		}
		_, gerr = tsvr.DoBusSvr([]byte(req))
		if gerr != nil {
			return gerr
		}
		if tsvr.RespCd != "00" && !strings.Contains(tsvr.RespMsg, "已存在") {
			return gerror.NewR(20050, nil, tsvr.RespMsg)
		}
		//time.Sleep(time.Second)
	}

	if cfg.SubAppid != "" {
		subAppLIst := strings.Split(cfg.SubAppid, ",")
		for _, subapp := range subAppLIst {
			req := T9007 + fmt.Sprintf("SubAppId=%s", subapp)
			tsvr, gerr := NewTranServices(tran.WepayCfg)
			if gerr != nil {
				return gerr
			}
			_, gerr = tsvr.DoBusSvr([]byte(req))
			if gerr != nil {
				return gerr
			}
			if tsvr.RespCd != "00" && !strings.Contains(tsvr.RespMsg, "已配置") {
				return gerror.NewR(20060, nil, tsvr.RespMsg)
			}
		}
	}

	return nil
}
