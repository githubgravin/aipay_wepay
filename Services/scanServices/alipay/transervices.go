package alipay

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"golib/gerror"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"
	"time"
	"unGateWay/Config"
	"unGateWay/Services/scanServices/scanCfg"
	"unGateWay/Services/scanServices/scanModel"
	"unGateWay/util"
)

type AlipayServices struct {
	ReqInfo IAlipayReq
	RspInfo IAlipayRsp
	reqMsg  *scanModel.TransMessage
	*scanCfg.AlipayCfg
}

func NewTranServices(cfg Config.ICfg) (*AlipayServices, gerror.IError) {
	var ok bool

	alipaySvr := AlipayServices{}

	alipaySvr.AlipayCfg, ok = cfg.(*scanCfg.AlipayCfg)
	if !ok {
		return nil, gerror.NewR(22001, nil, "非法的配置信息", cfg)
	}

	return &alipaySvr, nil
}

func (tran *AlipayServices) clone() *AlipayServices {
	alipaySvr := AlipayServices{}
	reqType := reflect.TypeOf(tran.ReqInfo).Elem()
	alipaySvr.ReqInfo = reflect.New(reqType).Interface().(IAlipayReq)
	rspType := reflect.TypeOf(tran.RspInfo).Elem()
	alipaySvr.RspInfo = reflect.New(rspType).Interface().(IAlipayRsp)

	util.CloneValue(tran.ReqInfo, alipaySvr.ReqInfo)
	util.CloneValue(tran.RspInfo, alipaySvr.RspInfo)
	alipaySvr.reqMsg = tran.reqMsg.Clone()
	alipaySvr.AlipayCfg = &scanCfg.AlipayCfg{}
	*alipaySvr.AlipayCfg = *tran.AlipayCfg

	return &alipaySvr
}

func (tran *AlipayServices) DoServices(req *scanModel.TransMessage) (*scanModel.TransMessage, gerror.IError) {

	var gerr gerror.IError
	tran.reqMsg = req

	tran.Debug("开始处理请求", req.Msg_body)

	switch req.MsgBody.Tran_cd {
	case "1131":
		gerr = tran.DoTrans(&Alipay_trade_pay{}, &AlipayRsp{}, false)
	case "1191":
		gerr = tran.DoTrans(&Alipay_trade_create{}, &AlipayRsp{}, true)
	case "2131":
		fallthrough
	case "3131":
		fallthrough
	case "3141":
		gerr = tran.DoTrans(&Alipay_trade_refund{}, &AlipayRsp{}, false)
	case "4131":
		gerr = tran.DoTrans(&Alipay_trade_close{}, &AlipayRsp{}, false)
	case "5131":
		gerr = tran.DoTrans(&Alipay_trade_query{}, &AlipayRsp{}, false)
	case "7131":
		gerr = tran.DoTrans(&Alipay_trade_pre_create{}, &AlipayRsp{}, true)
	default:
		return nil, gerror.NewR(1002, nil, "非法交易码[%s]", req.MsgBody.Tran_cd)
	}
	if gerr != nil {
		return nil, gerr
	}

	return tran.reqMsg, nil
}

func (tran *AlipayServices) DoTrans(req IAlipayReq, rsp IAlipayRsp, notiFlg bool) gerror.IError {

	var gerr gerror.IError

	tran.ReqInfo = req
	tran.RspInfo = rsp

	comReq := CommonRequest{}
	comReq.InitBase()
	comReq.Method = req.GetMethod()

	if notiFlg {
		comReq.Notify_url = tran.NotifyUrl
	}

	//cfg
	comReq.App_id = tran.AppId
	comReq.Timestamp = time.Now().Format("2006-01-02 15:04:05")

	gerr = req.InitRequest(tran, tran.reqMsg)
	if gerr != nil {
		return gerr
	}

	comReq.Biz_content, gerr = req.ToString()
	if gerr != nil {
		return gerr
	}

	respMsg, gerr := tran.PostVal(&comReq)
	if gerr != nil {
		return gerr
	}

	gerr = rsp.LoadResponse(respMsg, tran.reqMsg)
	if gerr != nil {
		return gerr
	}

	if tran.reqMsg != nil && tran.reqMsg.MsgBody != nil && tran.reqMsg.MsgBody.Qr_code_info != nil &&
		tran.reqMsg.MsgBody.Resp_cd == "W3" &&
		tran.reqMsg.MsgBody.Qr_code_info.Scance != "SYSTEM" && tran.reqMsg.MsgBody.Tran_cd == "1131" {
		selTran := tran.clone()
		selTran.reqMsg.MsgBody.Tran_cd = "5131"
		for i := 0; i < tran.QueryCnt; i++ {
			time.Sleep(time.Duration(tran.QueryInt) * time.Second)
			gerr = selTran.DoTrans(&Alipay_trade_query{}, &AlipayRsp{}, false)
			if gerr != nil {
				return gerr
			}
			if selTran.reqMsg.MsgBody.Resp_cd != "00" {
				tran.Debug("查询状态失败,继续查询", selTran.reqMsg.MsgBody.Resp_cd, selTran.reqMsg.MsgBody.Resp_msg)
				continue
			}
			if selTran.reqMsg.MsgBody.Orig_resp_cd != "00" {
				tran.Debug("查询成功，原交易未成功，继续查询",
					selTran.reqMsg.MsgBody.Orig_resp_cd, selTran.reqMsg.MsgBody.Orig_resp_msg)
				continue
			}
			//还原交易码
			selTran.reqMsg.MsgBody.Tran_cd = "1131"
			tran.reqMsg = selTran.reqMsg
			tran.reqMsg.MsgBody.Resp_cd = selTran.reqMsg.MsgBody.Orig_resp_cd
			tran.reqMsg.MsgBody.Resp_msg = selTran.reqMsg.MsgBody.Orig_resp_msg
			tran.reqMsg.MsgBody.Orig_resp_cd = ""
			tran.reqMsg.MsgBody.Orig_resp_msg = ""
			return nil
		}
		tran.Info("交易查询未明，撤销交易", tran.reqMsg.MsgBody.Order_id)
		//查询无结果，撤销
		//canTran := tran.clone()
		//canTran.reqMsg.MsgBody.Tran_cd = "4131"
		//canTran.reqMsg.MsgBody.Orig_order_id = tran.reqMsg.MsgBody.Order_id
		//for i := 0; i < tran.CancelCnt; i++ {
		//	time.Sleep(time.Duration(tran.QueryInt) * time.Second)
		//	gerr = canTran.DoTrans(&Alipay_trade_cancel{}, &AlipayRsp{}, false)
		//	if gerr != nil {
		//		return gerr
		//	}
		//	if canTran.reqMsg.MsgBody.Resp_cd != "00" {
		//		tran.Debug("撤消失败,继续撤销", canTran.reqMsg.MsgBody.Resp_cd, canTran.reqMsg.MsgBody.Resp_msg)
		//		continue
		//	}
		//	tran.reqMsg.MsgBody.Resp_cd = "S5"
		//	tran.reqMsg.MsgBody.Resp_msg = "交易失败已撤销"
		//	return nil
		//}
		tran.reqMsg.MsgBody.Resp_cd = "98"
		tran.reqMsg.MsgBody.Resp_msg = "交易状态不明，请稍后查询结果"
		return nil
	}

	return nil
}

func (tran *AlipayServices) PostVal(comReq *CommonRequest) (string, gerror.IError) {

	postVal := make(url.Values, 0)

	pt := reflect.TypeOf(comReq).Elem()
	pv := reflect.ValueOf(comReq).Elem()
	var signBuf bytes.Buffer
	for i := 0; i < pv.NumField(); i++ {
		val := string(UTF8tGBK([]byte(pv.Field(i).String())))
		//val := string(GBKtUTF8([]byte(pv.Field(i).String())))
		//val := string(pv.Field(i).String())
		if len(val) > 0 {
			nm := strings.ToLower(pt.Field(i).Name)
			postVal.Set(nm, val)
			signBuf.WriteString(nm + "=" + val + "&")
			tran.Debugf("%s[%s]", nm, GBKtUTF8(val))
		}
	}
	/*计算签名*/
	sign, gerr := RsaSignSha1Base64(tran.SignPrivKey, []byte(signBuf.Bytes()[:signBuf.Len()-1]))
	//sign, gerr := RsaSignSha1Base64(tran.SignPrivKey, []byte(postVal.Encode()))
	if gerr != nil {
		return "", gerr
	}
	postVal.Set("sign", sign)

	tran.Info("PostVal: 交易请求报文:", postVal.Encode())

	/*通信发送包*/
	tr := &http.Transport{
		Dial: (&net.Dialer{
			LocalAddr: tran.LocalAddr,
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: time.Second * time.Duration(tran.ReqTimeOut)}
	body := strings.NewReader(postVal.Encode())

	req, _ := http.NewRequest("POST", tran.RemoteURL, body)
	//req.Header.Set("Accept-Language", "zh")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;param=value")
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows;U;WindowsNT5.1;zh-CN;rv:1.8.1.11)Gecko/20071127 Firefox/2.0.0.11")

	out, _ := httputil.DumpRequestOut(req, true)

	tran.Debugf("http 请求包: %+v ", string(out))

	resp, err := client.Do(req)
	if err != nil {
		return "", gerror.NewR(20020, err, "POST:client.Do error; ")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", gerror.NewR(20030, nil, "POST:client.Do Status[%s] is not success; ", resp.StatusCode)
	}

	//读取应答
	data, _ := ioutil.ReadAll(resp.Body)
	res := GBKtUTF8(string(data))
	tran.Debugf("响应报文：[%v]\n-------------------\n", res)

	return res, nil
}

func (tran *AlipayServices) DoNotify(req []byte) ([]byte, gerror.IError) {

	qu, err := url.ParseQuery(string(req))
	if err != nil {
		return nil, gerror.NewR(15001, err, "解析报文失败", string(req))
	}
	reqMsg := scanModel.TransMessage{}
	reqMsg.MsgBody = &scanModel.TransParams{}
	reqMsg.MsgBody.Tran_cd = "6131"
	reqMsg.MsgBody.InsIdCd = tran.InsIdCd

	tran.Debug(qu.Get("trade_status"))
	if qu.Get("trade_status") == "TRADE_SUCCESS" {
		reqMsg.MsgBody.Orig_resp_cd = "00"
		reqMsg.MsgBody.Orig_resp_msg = "SUCCESS"
	} else {
		return nil, gerror.NewR(15010, nil, "收到未成功通知，不处理")
		//reqMsg.MsgBody.Orig_resp_cd = "E0"
		//reqMsg.MsgBody.Orig_resp_msg = "FAILED"
	}
	reqMsg.MsgBody.Sys_order_id = qu.Get("trade_no")
	reqMsg.MsgBody.Orig_sys_order_id = qu.Get("trade_no")
	reqMsg.MsgBody.Order_id = qu.Get("out_trade_no")
	reqMsg.MsgBody.Qr_code_info = &scanModel.QrCodeInfo{}
	reqMsg.MsgBody.Qr_code_info.Open_id = qu.Get("open_id")
	reqMsg.MsgBody.Qr_code_info.Pay_time = TimeConv(qu.Get("gmt_payment"))
	reqMsg.MsgBody.Tran_amt = Yuan2Points(qu.Get("total_amount"))
	reqMsg.MsgBody.Ma_chk_key = tran.InsIdCd + "PAY" + reqMsg.MsgBody.Order_id
	reqMsg.MsgBody.Mcht_nm = GBKtUTF8(qu.Get("subject"))
	reqMsg.MsgBody.Qr_code_info.Buyer_id = qu.Get("buyer_id")
	reqMsg.MsgBody.Qr_code_info.Buyer_user = qu.Get("buyer_logon_id")

	go tran.SendNoti(&reqMsg)

	return []byte("success"), nil
}
func (tran *AlipayServices) SendNoti(msg *scanModel.TransMessage) gerror.IError {

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

func (tran *AlipayServices) DoComm(request []byte, remoteAddr string) ([]byte, gerror.IError) {

	tran.Debug("remoteurl: ", remoteAddr)
	/*通信发送包*/
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: time.Second * time.Duration(tran.ReqTimeOut)}

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

func (tran *AlipayServices) DoBusSvr(req []byte) ([]byte, gerror.IError) {
	var gerr gerror.IError

	qu, err := url.ParseQuery(string(req))
	if err != nil {
		return nil, gerror.NewR(15001, err, "解析报文失败", string(req))
	}
	tran.Debug(qu)

	tranCd := qu.Get("Tran_cd")
	var mchtRsp AlipayRsp
	switch tranCd {
	case "9001":
		impReq := Ant_merchant_expand_indirect_create{}
		impReq.ExternalId = qu.Get("MchntCd")      /*本地商户号*/
		impReq.Name = qu.Get("MchntNm")            /*商户名称*/
		impReq.AliasName = qu.Get("MchntShortNm")  /*商户简称*/
		impReq.ServicePhone = qu.Get("MchntPhone") /*客服电话*/
		impReq.CategoryId = qu.Get("MchntCateId")  /*分类编目*/
		impReq.Memo = qu.Get("Memo")
		pid := qu.Get("PID")
		if len(pid) > 0 {
			impReq.Source = pid
		} else {
			impReq.Source = tran.SysServiceProviderID
		}
		//M2
		if qu.Get("ProCode") != "" {
			addr := &Address_info{}
			//addr.AddressInfo.Type = "BUSINESS_ADDRESS"
			addr.Province_code = qu.Get("ProCode")
			addr.City_code = qu.Get("CityCode")
			addr.District_code = qu.Get("DistCode")
			addr.Address = qu.Get("Memo")
			impReq.AddressInfo = append(make([]*Address_info, 0), addr)
		}
		gerr = tran.DoTrans(&impReq, &mchtRsp, false)
	case "9002":
		updReq := Ant_merchant_expand_indirect_modify{}
		updReq.SubMerchantId = qu.Get("AliMchntCd") /*支付宝商户号*/
		updReq.ExternalId = qu.Get("MchntCd")       /*本地商户号*/
		updReq.Name = qu.Get("MchntNm")             /*商户名称*/
		updReq.AliasName = qu.Get("MchntShortNm")   /*商户简称*/
		updReq.ServicePhone = qu.Get("MchntPhone")  /*客服电话*/
		updReq.CategoryId = qu.Get("MchntCateId")   /*分类编目*/
		pid := qu.Get("PID")
		if len(pid) > 0 {
			updReq.Source = pid
		} else {
			updReq.Source = tran.SysServiceProviderID
		}
		if qu.Get("ProCode") != "" {
			addr := &Address_info{}
			//addr.AddressInfo.Type = "BUSINESS_ADDRESS"
			addr.Province_code = qu.Get("ProCode")
			addr.City_code = qu.Get("CityCode")
			addr.District_code = qu.Get("DistCode")
			addr.Address = qu.Get("Memo")
			updReq.AddressInfo = append(make([]*Address_info, 0), addr)
		}
		gerr = tran.DoTrans(&updReq, &mchtRsp, false)
	case "9003":
		qry := Ant_merchant_expand_indirect_query{}
		qry.Sub_merchant_id = qu.Get("AliMchntCd")
		gerr = tran.DoTrans(&qry, &mchtRsp, false)
	case "9005":
		aliSett := SettFileReq{}
		aliSett.BillDate = qu.Get("BillDate")
		aliSett.BillType = qu.Get("BillType")
		gerr = tran.DoTrans(&aliSett, &mchtRsp, false)
	}
	if gerr != nil {
		return nil, gerr
	}

	msgRes := make(map[string]string, 0)
	switch tranCd {
	case "9001":
		msgRes["DstMchntCd"] = mchtRsp.MchtCrt.SubMerchantId
		msgRes["RetCd"] = RespConv(mchtRsp.MchtCrt.Code, "")
		msgRes["RetMsg"] = mchtRsp.MchtCrt.Msg + mchtRsp.MchtCrt.SubMsg
	case "9002":
		msgRes["DstMchntCd"] = mchtRsp.MchtUpd.SubMerchantId
		msgRes["RetCd"] = RespConv(mchtRsp.MchtUpd.Code, "")
		msgRes["RetMsg"] = mchtRsp.MchtUpd.Msg + mchtRsp.MchtUpd.SubMsg
	case "9005":
		msgRes["RetCd"] = RespConv(mchtRsp.BillDownrlQueryResponse.Code, "")
		msgRes["RetMsg"] = mchtRsp.BillDownrlQueryResponse.Msg
		//msgRes["DownUrl"] = mchtRsp.BillDownrlQueryResponse.BillDownloadURL
		if msgRes["RetCd"] == "00" {
			res, gerr := DownLoadFile(tran.FilePath, mchtRsp.BillDownrlQueryResponse.BillDownloadURL)
			if gerr != nil {
				return nil, gerr
			}
			msgRes["LocalFile"] = res
		}
	}

	res, err := json.Marshal(msgRes)
	if err != nil {
		return nil, gerror.NewR(17001, err, "生成应答失败")
	}

	return res, nil
}
