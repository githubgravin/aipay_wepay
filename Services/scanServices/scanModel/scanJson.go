package scanModel

import (
	"encoding/json"
	"golib/gerror"
)

type MessageHead struct {
	Encoding    string `json:"encoding"`
	Sign_method string `json:"sign_method"`
	Signature   string `json:"signature"`
	Version     string `json:"version"`
}

type commonParams struct {
	/*机构信息*/
	InsIdCd          string `json:"ins_id_cd,omitempty"`
	ChnInsIdCd       string `json:"chn_ins_id_cd,omitempty"` //发起系统机构号，标识密钥用。默认和InsIdCd相同
	Tran_cd          string `json:"tran_cd,omitempty"`
	Prod_cd          string `json:"prod_cd,omitempty"`
	Biz_cd           string `json:"biz_cd,omitempty"`
	Term_seq         string `json:"term_seq,omitempty"`
	Term_batch       string `json:"term_batch,omitempty"`
	Mcht_cd          string `json:"mcht_cd,omitempty"`
	Mcht_nm          string `json:"mcht_nm,omitempty"`
	Term_id          string `json:"term_id,omitempty"`
	Tran_dt_tm       string `json:"tran_dt_tm,omitempty"`
	Order_id         string `json:"order_id,omitempty"`
	Order_timeout    string `json:"order_timeout,omitempty"`
	Sys_order_id     string `json:"sys_order_id,omitempty"`
	Tran_order_id    string `json:"tran_order_id,omitempty"` //交易订单号，汇宜系统内唯一
	Acct_order_id    string `json:"acct_order_id,omitempty"` //账务订单号，同一笔账务对应交易相同
	Order_desc       string `json:"order_desc,omitempty"`
	Req_reserved     string `json:"req_reserved,omitempty"`
	Resp_cd          string `json:"resp_cd,omitempty"`
	Resp_msg         string `json:"resp_msg,omitempty"`
	ActiveCode       string `json:"active_code,omitempty"`
	Pri_acct_no      string `json:"pri_acct_no,omitempty"`
	Tran_amt         string `json:"tran_amt,omitempty"`
	Curr_cd          string `json:"curr_cd,omitempty"`
	Pre_auth_id      string `json:"pre_auth_id,omitempty"`
	Sett_dt          string `json:"sett_dt"`
	Iss_ins_id_cd    string `json:"iss_ins_id_cd"`
	Trans_in_acct_no string `json:"trans_in_acct_no,omitempty"`
}

type TransParams struct {
	commonParams
	Qr_code_info      *QrCodeInfo `json:"qr_code_info,omitempty"`
	Orig_sys_order_id string      `json:"orig_sys_order_id,omitempty"`
	Orig_order_id     string      `json:"orig_order_id,omitempty"`
	Orig_term_seq     string      `json:"orig_term_seq,omitempty"`
	Orig_resp_cd      string      `json:"orig_resp_cd,omitempty"`
	Orig_resp_msg     string      `json:"orig_resp_msg,omitempty"`
	Orig_trans_dt     string      `json:"orig_trans_dt,omitempty"`
	Orig_trans_dt_tm  string      `json:"orig_trans_dt_tm,omitempty"`
	Orig_tran_cd      string      `json:"orig_tran_cd,omitempty"`
	Orig_tran_nm      string      `json:"orig_tran_nm,omitempty"`
	Cancel_flg        string      `json:"cancel_flg,omitempty"`
	Orig_prod_cd      string      `json:"orig_prod_cd,omitempty"`
	Orig_biz_cd       string      `json:"orig_biz_cd,omitempty"`
	Ma_chk_key        string      `json:"ma_chk_key,omitempty"`
}

type QrCodeInfo struct {
	Auth_code string `json:"auth_code,omitempty"`
	Qr_type   string `json:"qr_type,omitempty"`
	Time_out  string `json:"time_out,omitempty"`
	Scance    string `json:"scance,omitempty"`
	//异步通知增加
	Sub_user_id string `json:"sub_user_id,omitempty"` //子用户id`
	Sub_app_id  string `json:"sub_app_id,omitempty"`  //子APPID`
	User_id     string `json:"User_id,omitempty"`     //用户ID`
	Noti_url    string `json:"noti_url,omitempty"`    //异步通知地址`
	Buyer_user  string `json:"buyer_user,omitempty"`  //用户登录帐号名称`
	Pay_time    string `json:"pay_time,omitempty"`    //订单完成时间`
	Subject     string `json:"subject,omitempty"`     //订单标题信息`
	Pay_bank    string `json:"pay_bank,omitempty"`    //支付渠道`
	Channel_id  string `json:"channel_id,omitempty"`  //渠道订单号`
	//Pay_order_id string `json:"pay_order_id,omitempty"`
	Qr_code       string `json:"qr_code,omitempty"`    //'订单支付ID`
	Wx_jsapi      string `json:"wx_jsapi,omitempty"`   //`微信字符串
	Open_id       string `json:"open_id,omitempty"`    //用户子ID`
	Buyer_id      string `json:"buyer_id,omitempty"`   //买家在支付宝的用户id`
	Cash_amt      string `json:"cash_amt,omitempty"`   //现金支付金额`
	Coupon_amt    string `json:"coupon_amt,omitempty"` //代金券金额`
	Trade_id      string `json:"trade_id,omitempty"`   //支付宝窗口支付`
	Goods_tag     string `json:"goods_tag,omitempty"`  //上传优惠信息`
	Order_ext1    string `json:"order_ext1,omitempty"` //订单扩展域1`
	Order_ext2    string `json:"order_ext2,omitempty"` //订单扩展域2`
	Store_id      string `json:"store_id,omitempty"`
	Refund_reason string `xml:"refund_reason,omitempty"`
}

type TransMessage struct {
	MessageHead
	Msg_body string       `json:"msg_body"`
	MsgBody  *TransParams `json:"-"`
}

func (t *TransMessage) Clone() *TransMessage {
	newTran := &TransMessage{}
	*newTran = *t
	newTran.MsgBody = &TransParams{}
	*newTran.MsgBody = *t.MsgBody
	return newTran
}

func UnPackReq(msg []byte) (*TransMessage, gerror.IError) {
	tranMsg := TransMessage{}

	err := json.Unmarshal([]byte(msg), &tranMsg)
	if err != nil {
		return nil, gerror.NewR(15001, err, "解析请求失败")
	}

	err = json.Unmarshal([]byte(tranMsg.Msg_body), &tranMsg.MsgBody)
	if err != nil {
		return nil, gerror.NewR(15001, err, "解析msgbody失败")
	}
	if tranMsg.MsgBody == nil {
		return nil, gerror.NewR(15001, nil, "msgbody不能为空")
	}

	return &tranMsg, nil
}

func (t *TransMessage) SetMsgBody() {
	btMsgBody, err := json.Marshal(t.MsgBody)
	if err != nil {
		t.Msg_body = "{}"
		return
	}
	t.Msg_body = string(btMsgBody)
}

func (t TransMessage) PackRsp() ([]byte, gerror.IError) {

	t.MsgBody.Tran_cd = t.MsgBody.Tran_cd[:3] + "2"

	t.SetMsgBody()

	res, err := json.Marshal(t)
	if err != nil {
		return nil, gerror.NewR(17001, err, "打包失败")
	}
	return res, nil
}

func (t TransMessage) PackReq() ([]byte, gerror.IError) {

	t.SetMsgBody()

	res, err := json.Marshal(t)
	if err != nil {
		return nil, gerror.NewR(17001, err, "打包失败")
	}
	return res, nil
}
