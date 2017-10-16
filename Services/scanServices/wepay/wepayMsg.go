package wepay

import "encoding/xml"

const (
	COM_SUCCESS        = "SUCCESS"
	COM_FAIL           = "FAIL"
	LOCAl_COM_FAIL     = "C0"
	TRAN_SUCCESS       = "SUCCESS"
	TRAN_FAIL          = "FAIL"
	LOCAL_BUSI_SUCCESS = "00"
	OTHER_BUSI_FAIL    = "E0"
)

var (
	RespCdMap map[string]RespCodeConv = map[string]RespCodeConv{
		/*查询应答状态转换*/             /*支付应答码转换*/
		"SUCCESS":               {"00", "成功"},
		"REFUND":                {"SF", "转入退款"},
		"NOTPAY":                {"W3", "未支付"},
		"CLOSED":                {"S5", "已关闭"},
		"REVOKED":               {"S8", "已撤销"},
		"USERPAYING":            {"W3", "用户支付中"},
		"PAYERROR":              {"W4", "支付失败"},
		"XML_FORMAT_ERROR":      {"F1", "XML格式错误"},
		"NOT_UTF8":              {"F2", "编码格式错误"},
		"REQUIRE_POST_METHOD":   {"F3", "请使用post方法"},
		"SIGNERROR":             {"F4", "签名错误"},
		"LACK_PARAMS":           {"F5", "缺少参数"},
		"SYSTEMERROR":           {"F6", "接口返回错误"},
		"INVALID_TRANSACTIONID": {"F7", "无效transaction_id"},
		"AUTH_CODE_ERROR":       {"Q1", "授权码参数错误"},
		"AUTH_CODE_INVALID":     {"Q2", "授权码检验错误"},
		"AUTHCODEEXPIRE":        {"Q3", "二维码已过期，请用户在微信上刷新后再试"},
		"BANKERROR":             {"Q4", "银行系统异常"},
		"BUYER_MISMATCH":        {"Q5", "支付帐号错误"},
		"MCHID_NOT_EXIST":       {"S1", "MCHID不存在"},
		"NOAUTH":                {"S2", "商户无此接口权限"},
		"NOTENOUGH":             {"S3", "余额不足"},
		"NOTSUPORTCARD":         {"S4", "不支持卡类型"},
		"ORDERCLOSED":           {"S5", "订单已关闭"},
		"ORDERNOTEXIST":         {"S6", "此交易订单号不存在"},
		"ORDERPAID":             {"S7", "商户订单已支付"},
		"ORDERREVERSED":         {"S8", "订单已撤销"},
		"OUT_TRADE_NO_USED":     {"S9", "商户订单号重复"},
		"PARAM_ERROR":           {"SA", "参数错误"},
		"POST_DATA_EMPTY":       {"SB", "post数据为空"},
		"REVERSE_EXPIRE":        {"SC", "订单无法撤销"},
		"APPID_MCHID_NOT_MATCH": {"SD", "appid和mch_id不匹配"},
		"APPID_NOT_EXIST":       {"SE", "APPID不存在"},
		"USER_ACCOUNT_ABNORMAL": {"W2", "退款请求失败"},
		"TRADE_STATE_ERROR":     {"SG", "订单状态错误"},
	}
)

type RespCodeConv struct {
	Code string
	Desc string
}

type WxJsapi struct {
	APPID     string `json:"appId,omitempty"`     /*公众号ID*/
	TimeStamp string `json:"timeStamp,omitempty"` /*时间戳*/
	NonceStr  string `json:"nonceStr,omitempty"`  /*随机字符串*/
	Package   string `json:"package,omitempty"`   /*订单详情扩展字符串*/
	SignType  string `json:"signType,omitempty"`  /*签名方式*/
	PaySign   string `json:"paySign,omitempty"`   /*签名*/
}

/*请求报文*/
type Request struct {
	XMLName        xml.Name `xml:"xml"`
	InterTranCode  string   `xml:"-"`
	APPId          string   `xml:"appid,omitempty"`            /*公众账号ID*/
	MchId          string   `xml:"mch_id,omitempty"`           /*商户号	*/
	SubAPPId       string   `xml:"sub_appid,omitempty"`        /*子商户公众账号ID*/
	SubMchId       string   `xml:"sub_mch_id,omitempty"`       /*子商户号	*/
	DeviceInfo     string   `xml:"device_info,omitempty"`      /*设备号	*/
	NonceStr       string   `xml:"nonce_str,omitempty"`        /*随机字符串*/
	Sign           string   `xml:"sign,omitempty"`             /*签名*/
	Body           string   `xml:"body,omitempty"`             /*商品描述*/
	Detail         string   `xml:"detail,omitempty,cddata"`    /*商品详情*/
	Attach         string   `xml:"attach,omitempty"`           /*附加数据*/
	OutTradeNo     string   `xml:"out_trade_no,omitempty"`     /*商户订单号*/
	TotalFee       int      `xml:"total_fee,omitempty"`        /*总金额*/
	FeeType        string   `xml:"fee_type,omitempty"`         /*货币类型*/
	SpBillCreateIP string   `xml:"spbill_create_ip,omitempty"` /*终端IP*/
	TimeStart      string   `xml:"time_start,omitempty"`       /*订单开始时间*/
	TimeExpire     string   `xml:"time_expire,omitempty"`      /*过期时间*/
	ProductId      string   `xml:"product_id,omitempty"`       /*商品ID*/
	GoodsTag       string   `xml:"goods_tag,omitempty"`        /*商品标记*/
	NotifyURL      string   `xml:"notify_url,omitempty"`       /*通知地址*/
	TradeType      string   `xml:"trade_type,omitempty"`       /*交易类型*/
	LimitPay       string   `xml:"limit_pay,omitempty"`        /*指定支付方式*/
	OpenId         string   `xml:"openid,omitempty"`           /*用户标识*/
	SubOpenId      string   `xml:"sub_openid,omitempty"`       /*用户子标识*/
	AuthCode       string   `xml:"auth_code,omitempty"`        /*授权码	*/
	TransactionId  string   `xml:"transaction_id,omitempty"`   /*微信订单号*/
	RefundId       string   `xml:"out_refund_no,omitempty"`    /*退款单号*/
	RefundFee      int      `xml:"refund_fee,omitempty"`       /*退款金额*/
	OpUserId       string   `xml:"op_user_id,omitempty"`       /*操作员号*/

	MchName      string `xml:"merchant_name,omitempty"`      /*商户名*/
	MchShtName   string `xml:"merchant_shortname,omitempty"` /*商户简称*/
	ServicePhone string `xml:"service_phone,omitempty"`      /* 客服电话*/
	Contact      string `xml:"contact,omitempty"`            /*联系人*/
	ContactPhone string `xml:"contact_phone,omitempty"`      /*联系电话*/
	ContactMail  string `xml:"contact_email,omitempty"`      /*联系邮箱*/
	Business     string `xml:"business,omitempty"`           /*经营类目*/
	MchRemark    string `xml:"merchant_remark,omitempty"`    /*商户备注（用于唯一性确定要素）*/
	ChannelId    string `xml:"channel_id,omitempty"`         /*渠道号*/
	PageIndex    string `xml:"page_index,omitempty"`         /* 页码*/
	PageSize     string `xml:"page_size,omitempty"`          /* 每页个数*/

	BillType string `xml:"bill_type,omitempty"` /*账单类型*/
	BillDate string `xml:"bill_date,omitempty"` /*账单日期*/

	JsapiPath      string `xml:"jsapi_path,omitempty"`      /*子商户公众账号 JSAPI 支付授权目录*/
	SubscribeAppid string `xml:"subscribe_appid,omitempty"` /*子商户推荐关注公众账号APPID*/
}

/*应答报文*/
type Response struct {
	XMLName       xml.Name //`xml:"xml"`
	InterTranCode string   `xml:"-"`
	ReturnCode    string   `xml:"return_code,cddata"`      /*通信应答码*/
	ReturnMsg     string   `xml:"return_msg,cddata"`       /*返回信息*/
	ResultCode    string   `xml:"result_code,cddata"`      /*业务应答码*/
	ResultMsg     string   `xml:"result_msg,cddata"`       /*业务错误信息 */
	ErrCode       string   `xml:"err_code,cddata"`         /*错误代码*/
	ErrCodeDes    string   `xml:"err_code_des,cddata"`     /*错误描述*/
	TradeState    string   `xml:"trade_state,cddata"`      /*交易状态*/
	TradeStateDes string   `xml:"trade_state_desc,cddata"` /*交易状态描述*/

	APPId             string    `xml:"appid,cddata"`                /*公众账号ID*/
	MchId             string    `xml:"mch_id,cddata"`               /*商户号*/
	SubAPPId          string    `xml:"sub_appid,cddata"`            /*子商户公众账号ID*/
	SubMchId          string    `xml:"sub_mch_id,cddata"`           /*子商户号*/
	DeviceInfo        string    `xml:"device_info,cddata"`          /*设备号*/
	NonceStr          string    `xml:"nonce_str,cddata"`            /*随机字符串*/
	Sign              string    `xml:"sign,cddata"`                 /*签名*/
	OpenId            string    `xml:"openid,cddata"`               /*错误描述*/
	IsSubscribe       string    `xml:"is_subscribe,cddata"`         /*is_subscribe*/
	SubOpenId         string    `xml:"sub_openid,cddata"`           /*用户子标识*/
	SubIsSubscribe    string    `xml:"sub_is_subscribe,cddata"`     /*是否关注子公众账号*/
	TradeType         string    `xml:"trade_type,cddata"`           /*交易类型*/
	PrePayId          string    `xml:"prepay_id,cddata"`            /*预支付交易会话标识*/
	CodeURL           string    `xml:"code_url,cddata"`             /*二维码链接*/
	BankType          string    `xml:"bank_type,cddata"`            /*付款银行*/
	FeeType           string    `xml:"fee_type,cddata"`             /*货币类型*/
	TotalFee          string    `xml:"total_fee,cddata"`            /*总金额*/
	CashFeeType       string    `xml:"cash_fee_type,cddata"`        /*现金支付货币类型*/
	CashFee           string    `xml:"cash_fee,cddata"`             /*现金支付金额*/
	SettlTotalFee     string    `xml:"settlement_total_fee,cddata"` /*应结订单金额*/
	TransactionId     string    `xml:"transaction_id,cddata"`       /*微信订单号*/
	OutTradeNo        string    `xml:"out_trade_no,cddata"`         /*外部订单号*/
	Detail            string    `xml:"detail,cddata"`               /*商品详情*/
	Attach            string    `xml:"attach,cddata"`               /*附加数据*/
	TimeEnd           string    `xml:"time_end,cddata"`             /*关注公众号*/
	RefundId          string    `xml:"refund_id,cddata"`            /*退款单号*/
	OutRefundNo       string    `xml:"out_refund_no,cddata"`        /*请求退款订单号*/
	RefundFee         string    `xml:"refund_fee"`                  /*退款金额*/
	CouponFee         string    `xml:"coupon_fee,cddata"`           /*代金券金额*/
	CashRefundFee     string    `xml:"cash_refund_fee"`             /*现金退款金额*/
	CouponRefundCount string    `xml:"coupon_refund_count"`         /*代金券退款次数*/
	CouponRefundFee   string    `xml:"coupon_refund_fee"`           /*代金券退款金额*/
	ReCall            string    `xml:"recall,cddata"`               /*是否重试*/
	Total             string    `xml:"total"`                       /*总记录数*/
	SubMchInfs        []MchtInf `xml:"mchinfo"`                     /*子商户列表*/
	ChanMchId         string    `xml:"channel_mch_id"`              /*渠道商商户号*/
	JsapiPathList     string    `xml:"jsapi_path_list"`             /*子商户公众号支付域名列表*/
	AppidConfList     string    `xml:"appid_config_list"`           /*特约商户 APPID 配置列表*/
}

type MchtInf struct {
	MchId        string `xml:"mch_id,omitempty"`             /*商户号*/
	MchName      string `xml:"merchant_name,omitempty"`      /*商户名*/
	MchShtName   string `xml:"merchant_shortname,omitempty"` /*商户简称*/
	ServicePhone string `xml:"service_phone,omitempty"`      /* 客服电话*/
	Contact      string `xml:"contact,omitempty"`            /*联系人*/
	ContactPhone string `xml:"contact_phone,omitempty"`      /*联系电话*/
	ContactMail  string `xml:"contact_email,omitempty"`      /*联系邮箱*/
	Business     string `xml:"business,omitempty"`           /*经营类目*/
	MchRemark    string `xml:"merchant_remark,omitempty"`    /*商户备注（用于唯一性确定要素）*/
}
