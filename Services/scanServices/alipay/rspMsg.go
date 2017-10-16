package alipay

import (
	"encoding/json"
	"golib/gerror"
	"unGateWay/Services/scanServices/scanModel"
)

const ALI_RESP_SUCCESS = "10000"
const ALI_RESP_EXCEPTION = "20000"
const ALI_RESP_TRANING = "10003"
const ALI_RESP_FAILED = "40004"
const QUERY_NUM = 11
const CANCEL_NUM = 5

/*阿里失败应答码转换*/
var ALI_ERR_CD_CONV map[string]string = map[string]string{
	"ACQ.SYSTEM_ERROR":                           "96",
	"ACQ.INVALID_PARAMETER":                      "96",
	"ACQ.TRADE_NOT_EXIST":                        "25",
	"TRADE_CLOSED":                               "C0",
	"ACQ.SELLER_BALANCE_NOT_ENOUGH":              "61",
	"ACQ.REFUND_AMT_NOT_EQUAL_TOTAL":             "61",
	"ACQ.REASON_TRADE_BEEN_FREEZEN":              "61",
	"ACQ.DISCORDANT_REPEAT_REQUEST":              "25",
	"ACQ.REASON_TRADE_REFUND_FEE_ERR":            "13",
	"ACQ.TRADE_NOT_ALLOW_REFUND":                 "40",
	"ACQ.ACCESS_FORBIDDEN":                       "40",
	"ACQ.EXIST_FORBIDDEN_WORD":                   "30",
	"ACQ.PARTNER_ERROR":                          "96",
	"ACQ.TOTAL_FEE_EXCEED":                       "13",
	"ACQ.PAYMENT_AUTH_CODE_INVALID":              "05",
	"ACQ.CONTEXT_INCONSISTENT":                   "A0",
	"ACQ.BUYER_BALANCE_NOT_ENOUGH":               "51",
	"ACQ.BUYER_BANKCARD_BALANCE_NOT_ENOUGH":      "51",
	"ACQ.ERROR_BALANCE_PAYMENT_DISABLE":          "52",
	"ACQ.BUYER_SELLER_EQUAL":                     "05",
	"ACQ.TRADE_BUYER_NOT_MATCH":                  "05",
	"ACQ.BUYER_ENABLE_STATUS_FORBID":             "05",
	"ACQ.PULL_MOBILE_CASHIER_FAIL":               "96",
	"ACQ.MOBILE_PAYMENT_SWITCH_OFF":              "05",
	"ACQ.PAYMENT_FAIL":                           "06",
	"ACQ.BUYER_PAYMENT_AMOUNT_DAY_LIMIT_ERROR":   "61",
	"ACQ.BEYOND_PAY_RESTRICTION":                 "03",
	"ACQ.BEYOND_PER_RECEIPT_RESTRICTION":         "03",
	"ACQ.BUYER_PAYMENT_AMOUNT_MONTH_LIMIT_ERROR": "61",
	"ACQ.SELLER_BEEN_BLOCKED":                    "03",
	"ACQ.ERROR_BUYER_CERTIFY_LEVEL_LIMIT":        "05",
	"ACQ.PAYMENT_REQUEST_HAS_RISK":               "01",
	"ACQ.NO_PAYMENT_INSTRUMENTS_AVAILABLE":       "01",
	"ACQ.USER_FACE_PAYMENT_SWITCH_OFF":           "05",
	"ACQ.INVALID_STORE_ID":                       "03",
	"ACQ.TRADE_STATUS_ERROR":                     "C2",
	"ACQ.TRADE_HAS_FINISHED":                     "C1",
	"ACQ.TRADE_HAS_CLOSE":                        "C3",
	"ACQ.TRADE_HAS_SUCCESS":                      "C4",
}

type AlipayRsp struct {
	TradePayResponse        AlipayTradePayResponse       `json:"alipay_trade_pay_response"`
	TradeQueryResponse      AlipayTradeQueryResponse     `json:"alipay_trade_query_response"`
	TradeRefundResponse     AlipayTradeRefundResponse    `json:"alipay_trade_refund_response"`
	TradeCancelResponse     AlipayTradeCancelResponse    `json:"alipay_trade_cancel_response"`
	TradePrecreateResponse  AlipayTradePrecreateResponse `json:"alipay_trade_precreate_response"`
	TradeCreateResponse     AlipayTradeCreateResponse    `json:"alipay_trade_create_response"`
	TradeCloseResponse      AlipayTradeCloseResponse     `json:"alipay_trade_close_response"`
	MchtCrt                 *MchtRsp                     `json:"ant_merchant_expand_indirect_create_response"`
	MchtUpd                 *MchtRsp                     `json:"ant_merchant_expand_indirect_modify_response"`
	NullRsp                 *NullRsp                     `json:"null_response"`
	BillDownrlQueryResponse BillDownloadurlQueryResponse `json:"alipay_data_dataservice_bill_downloadurl_query_response"`
	Sign                    string                       `json:"sign"`
}

func (t *AlipayRsp) LoadResponse(rspMsg string, tmsg *scanModel.TransMessage) gerror.IError {
	err := json.Unmarshal([]byte(rspMsg), t)
	if err != nil {
		return gerror.NewR(14001, err, "解析应答失败")
	}

	if tmsg == nil {
		return nil
	}
	switch tmsg.MsgBody.Tran_cd {
	case "1131":
		tmsg.MsgBody.Resp_cd = RespConv(t.TradePayResponse.Code, t.TradePayResponse.SubCode)
		tmsg.MsgBody.Resp_msg = t.TradePayResponse.Msg + t.TradePayResponse.SubMsg
		tmsg.MsgBody.Sys_order_id = t.TradePayResponse.TradeNo
		tmsg.MsgBody.Qr_code_info.Buyer_id = t.TradePayResponse.BuyerUserID
		tmsg.MsgBody.Qr_code_info.Buyer_user = t.TradePayResponse.BuyerLogonID
		tmsg.MsgBody.Qr_code_info.Open_id = t.TradePayResponse.OpenID
		tmsg.MsgBody.Qr_code_info.Pay_time = TimeConv(t.TradePayResponse.GmtPayment)
		str, _ := json.Marshal(t.TradePayResponse.FundBillList)
		tmsg.MsgBody.Qr_code_info.Pay_bank = string(str)
	case "1191":
		tmsg.MsgBody.Resp_cd = RespConv(t.TradeCreateResponse.Code, t.TradeCreateResponse.SubCode)
		tmsg.MsgBody.Resp_msg = t.TradeCreateResponse.Msg + t.TradeCreateResponse.SubMsg
		tmsg.MsgBody.Sys_order_id = t.TradeCreateResponse.TradeNo
	case "5131":
		tmsg.MsgBody.Resp_cd = RespConv(t.TradeQueryResponse.Code, t.TradeQueryResponse.SubCode)
		tmsg.MsgBody.Resp_msg = t.TradeQueryResponse.Msg + t.TradeQueryResponse.SubMsg
		tmsg.MsgBody.Orig_sys_order_id = t.TradeQueryResponse.TradeNo
		tmsg.MsgBody.Orig_order_id = t.TradeQueryResponse.OutTradeNo
		tmsg.MsgBody.Qr_code_info = &scanModel.QrCodeInfo{}
		tmsg.MsgBody.Qr_code_info.Buyer_id = t.TradeQueryResponse.BuyerUserID
		tmsg.MsgBody.Qr_code_info.Buyer_user = t.TradeQueryResponse.BuyerLogonID
		tmsg.MsgBody.Qr_code_info.Open_id = t.TradeQueryResponse.OpenID
		tmsg.MsgBody.Qr_code_info.Pay_time = TimeConv(t.TradeQueryResponse.SendPayDate)
		str, _ := json.Marshal(t.TradeQueryResponse.FundBillList)
		tmsg.MsgBody.Qr_code_info.Pay_bank = string(str)
		if tmsg.MsgBody.Resp_cd == "00" {
			tmsg.MsgBody.Orig_resp_cd = TradeStatConv(t.TradeQueryResponse.TradeStatus, t.TradeQueryResponse.SubCode)
		}
	case "4131":
		tmsg.MsgBody.Resp_cd = RespConv(t.TradeCancelResponse.Code, t.TradeCancelResponse.SubCode)
		tmsg.MsgBody.Resp_msg = t.TradeCancelResponse.Msg + t.TradeCancelResponse.SubMsg
		tmsg.MsgBody.Orig_sys_order_id = t.TradeCancelResponse.TradeNo
		tmsg.MsgBody.Orig_order_id = t.TradeCancelResponse.OutTradeNo
	case "2131":
		fallthrough
	case "3131":
		fallthrough
	case "3141":
		tmsg.MsgBody.Resp_cd = RespConv(t.TradeRefundResponse.Code, t.TradeRefundResponse.SubCode)
		tmsg.MsgBody.Resp_msg = t.TradeRefundResponse.SubMsg
		tmsg.MsgBody.Sys_order_id = t.TradeRefundResponse.TradeNo
		tmsg.MsgBody.Qr_code_info = &scanModel.QrCodeInfo{}
		tmsg.MsgBody.Qr_code_info.Buyer_id = t.TradeRefundResponse.BuyerUserID
		tmsg.MsgBody.Qr_code_info.Buyer_user = t.TradeRefundResponse.BuyerLogonID
		tmsg.MsgBody.Qr_code_info.Open_id = t.TradeRefundResponse.OpenID
		tmsg.MsgBody.Qr_code_info.Pay_time = TimeConv(t.TradeRefundResponse.GmtRefundPay)
	case "7131":
		tmsg.MsgBody.Resp_cd = RespConv(t.TradePrecreateResponse.Code, t.TradePrecreateResponse.SubCode)
		tmsg.MsgBody.Resp_msg = t.TradePrecreateResponse.Msg + t.TradePrecreateResponse.SubMsg
		tmsg.MsgBody.Qr_code_info = &scanModel.QrCodeInfo{}
		tmsg.MsgBody.Qr_code_info.Qr_code = t.TradePrecreateResponse.QrCode
	case "9001":
	default:
		return gerror.NewR(14010, nil, "不支持的交易码", tmsg.MsgBody.Tran_cd)
	}

	return nil
}

type FundBill struct {
	Amount      string `json:"amount,omitempty"`
	FundChannel string `json:"fund_channel,omitempty"`
	RealAmount  string `json:"real_amount,omitempty"`
}

type NullRsp struct {
	Code    string `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	SubCode string `json:"sub_code,omitempty"`
	SubMsg  string `json:"sub_msg,omitempty"`
}

type AlipayTradePayResponse struct {
	Code                string     `json:"code,omitempty"`
	Msg                 string     `json:"msg,omitempty"`
	SubCode             string     `json:"sub_code,omitempty"`
	SubMsg              string     `json:"sub_msg,omitempty"`
	BuyerLogonID        string     `json:"buyer_logon_id,omitempty"`
	BuyerPayAmount      string     `json:"buyer_pay_amount,omitempty"`
	BuyerUserID         string     `json:"buyer_user_id,omitempty"`
	FundBillList        []FundBill `json:"fund_bill_list,omitempty"`
	GmtPayment          string     `json:"gmt_payment,omitempty"`
	InvoiceAmount       string     `json:"invoice_amount,omitempty"`
	OpenID              string     `json:"open_id,omitempty"`
	OutTradeNo          string     `json:"out_trade_no,omitempty"`
	PointAmount         string     `json:"point_amount,omitempty"`
	ReceiptAmount       string     `json:"receipt_amount,omitempty"`
	TotalAmount         string     `json:"total_amount,omitempty"`
	TradeNo             string     `json:"trade_no,omitempty"`
	CardBalance         string     `json:"card_balance,omitempty"`
	DiscountGoodsDetail string     `json:"discount_goods_detail,omitempty"`
	StoreName           string     `json:"store_name,omitempty"`
}

type AlipayTradeQueryResponse struct {
	AlipayStoreID       string     `json:"alipay_store_id,omitempty"`
	BuyerLogonID        string     `json:"buyer_logon_id,omitempty"`
	BuyerPayAmount      string     `json:"buyer_pay_amount,omitempty"`
	BuyerUserID         string     `json:"buyer_user_id,omitempty"`
	Code                string     `json:"code,omitempty"`
	SubCode             string     `json:"sub_code,omitempty"`
	DiscountGoodsDetail string     `json:"discount_goods_detail,omitempty"`
	FundBillList        []FundBill `json:"fund_bill_list,omitempty"`
	IndustrySepcDetail  string     `json:"industry_sepc_detail,omitempty"`
	InvoiceAmount       string     `json:"invoice_amount,omitempty"`
	Msg                 string     `json:"msg,omitempty"`
	SubMsg              string     `json:"sub_msg,omitempty"`
	OpenID              string     `json:"open_id,omitempty"`
	OutTradeNo          string     `json:"out_trade_no,omitempty"`
	PointAmount         string     `json:"point_amount,omitempty"`
	ReceiptAmount       string     `json:"receipt_amount,omitempty"`
	SendPayDate         string     `json:"send_pay_date,omitempty"`
	StoreID             string     `json:"store_id,omitempty"`
	StoreName           string     `json:"store_name,omitempty"`
	TerminalID          string     `json:"terminal_id,omitempty"`
	TotalAmount         string     `json:"total_amount,omitempty"`
	TradeNo             string     `json:"trade_no,omitempty"`
	TradeStatus         string     `json:"trade_status,omitempty"`
}

type refundDetailItemList struct {
	Amount      string `json:"amount,omitempty"`
	FundChannel string `json:"fund_channel,omitempty"`
}

type AlipayTradeRefundResponse struct {
	Code                 string                 `json:"code,omitempty"`
	Msg                  string                 `json:"msg,omitempty"`
	SubCode              string                 `json:"sub_code,omitempty"`
	SubMsg               string                 `json:"sub_msg,omitempty"`
	BuyerLogonID         string                 `json:"buyer_logon_id,omitempty"`
	BuyerUserID          string                 `json:"buyer_user_id,omitempty"`
	FundChange           string                 `json:"fund_change,omitempty"`
	GmtRefundPay         string                 `json:"gmt_refund_pay,omitempty"`
	OpenID               string                 `json:"open_id,omitempty"`
	OutTradeNo           string                 `json:"out_trade_no,omitempty"`
	RefundDetailItemList []refundDetailItemList `json:"refund_detail_item_list,omitempty"`
	RefundFee            string                 `json:"refund_fee,omitempty"`
	SendBackFee          string                 `json:"send_back_fee,omitempty"`
	TradeNo              string                 `json:"trade_no,omitempty"`
}

type AlipayTradeCancelResponse struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	SubCode    string `json:"sub_code"`
	SubMsg     string `json:"sub_msg"`
	OutTradeNo string `json:"out_trade_no"`
	TradeNo    string `json:"trade_no"`
	RetryFlag  string `json:"retry_flag"`
	Action     string `json:"action"`
}

type AlipayTradePrecreateResponse struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	SubCode    string `json:"sub_code"`
	SubMsg     string `json:"sub_msg"`
	OutTradeNo string `json:"out_trade_no"`
	QrCode     string `json:"qr_code"`
}

type AlipayTradeCreateResponse struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	SubCode    string `json:"sub_code"`
	SubMsg     string `json:"sub_msg"`
	OutTradeNo string `json:"out_trade_no"`
	TradeNo    string `json:"trade_no"`
}

type AlipayTradeCloseResponse struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	SubCode    string `json:"sub_code"`
	SubMsg     string `json:"sub_msg"`
	OutTradeNo string `json:"out_trade_no"`
	TradeNo    string `json:"trade_no"`
}

type MchtRsp struct {
	Code          string `json:"code"`
	Msg           string `json:"msg"`
	SubCode       string `json:"sub_code"`
	SubMsg        string `json:"sub_msg"`
	SubMerchantId string `json:"sub_merchant_id"`
}

type BillDownloadurlQueryResponse struct {
	BillDownloadURL string `json:"bill_download_url"`
	Code            string `json:"code"`
	Msg             string `json:"msg"`
}
