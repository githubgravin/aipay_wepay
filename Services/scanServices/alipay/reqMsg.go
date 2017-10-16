package alipay

import (
	"encoding/json"
	"fmt"
	"golib/gerror"
	"strconv"
	"unGateWay/Services/scanServices/scanModel"
)

type goodsDetail struct {
	Goods_id        string `json:"goods_id,omitempty"`
	Alipay_goods_id string `json:"alipay_goods_id,omitempty"`
	Goods_name      string `json:"goods_name,omitempty"`
	Quantity        string `json:"quantity,omitempty"`
	Price           string `json:"price,omitempty"`
	Goods_category  string `json:"goods_category,omitempty"`
	Body            string `json:"body,omitempty"`
	Show_url        string `json:"show_url,omitempty"`
}

type extendParams struct {
	Sys_service_provider_id string `json:"sys_service_provider_id,omitempty"`
	Hb_fq_num               string `json:"hb_fq_num,omitempty"`
	Hb_fq_seller_percent    string `json:"hb_fq_seller_percent,omitempty"`
}

type royaltyInfo struct {
	Royalty_type         string               `json:"royalty_type,omitempty"`
	Royalty_detail_infos []royaltyDetailInfos `json:"royalty_detail_infos,omitempty"`
}
type royaltyDetailInfos struct {
	Serial_no         string `json:"serial_no,omitempty"`
	Trans_in_type     string `json:"trans_in_type,omitempty"`
	Batch_no          string `json:"batch_no,omitempty"`
	Out_relation_id   string `json:"out_relation_id,omitempty"`
	Trans_out_type    string `json:"trans_out_type,omitempty"`
	Trans_out         string `json:"trans_out,omitempty"`
	Trans_in          string `json:"trans_in,omitempty"`
	Amount            string `json:"amount,omitempty"`
	Desc              string `json:"desc,omitempty"`
	Amount_percentage string `json:"amount_percentage,omitempty"`
}
type subMerchant struct {
	Merchant_id string `json:"merchant_id,omitempty"`
}

type Alipay_trade_pay struct {
	Out_trade_no          string        `json:"out_trade_no,omitempty"`
	Scene                 string        `json:"scene,omitempty"`
	Auth_code             string        `json:"auth_code,omitempty"`
	Subject               string        `json:"subject,omitempty"`
	Seller_id             string        `json:"seller_id,omitempty"`
	Total_amount          string        `json:"total_amount,omitempty"`
	Discountable_amount   string        `json:"discountable_amount,omitempty"`
	Undiscountable_amount string        `json:"undiscountable_amount,omitempty"`
	Body                  string        `json:"body,omitempty"`
	Goods_detail          []goodsDetail `json:"goods_detail,omitempty"`
	Operator_id           string        `json:"operator_id,omitempty"`
	Store_id              string        `json:"store_id,omitempty"`
	Terminal_id           string        `json:"terminal_id,omitempty"`
	Alipay_store_id       string        `json:"alipay_store_id,omitempty"`
	Extend_params         *extendParams `json:"extend_params,omitempty"`
	Timeout_express       string        `json:"timeout_express,omitempty"`
	Royalty_info          *royaltyInfo  `json:"royalty_info,omitempty"`
	Sub_merchant          *subMerchant  `json:"sub_merchant,omitempty"`
	Auth_no               string        `json:"auth_no,omitempty"`
}

func (t Alipay_trade_pay) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Alipay_trade_pay 打包失败")
	}
	return string(res), nil
}

func (t Alipay_trade_pay) GetMethod() string {
	return "alipay.trade.pay"
}

func (t *Alipay_trade_pay) InitRequest(tran *AlipayServices, msg *scanModel.TransMessage) gerror.IError {
	var gerr gerror.IError

	t.Sub_merchant = &subMerchant{}
	t.Sub_merchant.Merchant_id = msg.MsgBody.Mcht_cd
	t.Out_trade_no = msg.MsgBody.Order_id
	t.Scene = "bar_code"
	t.Auth_code = msg.MsgBody.Qr_code_info.Auth_code
	t.Total_amount, gerr = DiveAmt(msg.MsgBody.Tran_amt)
	if gerr != nil {
		return gerr
	}
	t.Subject = msg.MsgBody.Mcht_nm
	if svrpid := tran.GetSvrPid(msg.MsgBody.InsIdCd); svrpid != "" {
		t.Extend_params = &extendParams{}
		t.Extend_params.Sys_service_provider_id = svrpid
	}
	if msg.MsgBody.Qr_code_info.Time_out != "" {
		secNum, err := strconv.Atoi(msg.MsgBody.Qr_code_info.Time_out)
		secNum = secNum / 60
		if err != nil {
			t.Timeout_express = tran.TranTimeOut
		} else if secNum == 0 {
			t.Timeout_express = "1m"
		} else {
			t.Timeout_express = fmt.Sprintf("%dm", secNum)
		}
	} else {
		t.Timeout_express = tran.TranTimeOut
	}
	t.Store_id = msg.MsgBody.Qr_code_info.Store_id
	t.Terminal_id = msg.MsgBody.Term_id

	return nil
}

type Alipay_trade_query struct {
	Out_trade_no string `json:"out_trade_no,omitempty"`
	Trade_no     string `json:"trade_no,omitempty"`
}

func (t Alipay_trade_query) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Alipay_trade_query 打包失败")
	}
	return string(res), nil
}

func (t Alipay_trade_query) GetMethod() string {
	return "alipay.trade.query"
}

func (t *Alipay_trade_query) InitRequest(tran *AlipayServices, msg *scanModel.TransMessage) gerror.IError {
	t.Trade_no = msg.MsgBody.Orig_sys_order_id
	t.Out_trade_no = msg.MsgBody.Order_id
	return nil
}

type Alipay_trade_refund struct {
	Out_trade_no   string `json:"out_trade_no,omitempty"`
	Trade_no       string `json:"trade_no,omitempty"`
	Refund_amount  string `json:"refund_amount,omitempty"`
	Refund_reason  string `json:"refund_reason,omitempty"`
	Out_request_no string `json:"out_request_no,omitempty"`
	Operator_id    string `json:"operator_id,omitempty"`
	Store_id       string `json:"store_id,omitempty"`
	Terminal_id    string `json:"terminal_id,omitempty"`
}

func (t Alipay_trade_refund) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Alipay_trade_refund 打包失败")
	}
	return string(res), nil
}
func (t Alipay_trade_refund) GetMethod() string {
	return "alipay.trade.refund"
}

func (t *Alipay_trade_refund) InitRequest(tran *AlipayServices, msg *scanModel.TransMessage) gerror.IError {
	var gerr gerror.IError

	t.Trade_no = msg.MsgBody.Orig_sys_order_id
	t.Out_trade_no = msg.MsgBody.Orig_order_id
	t.Refund_amount, gerr = DiveAmt(msg.MsgBody.Tran_amt)
	if gerr != nil {
		return gerr
	}
	t.Refund_reason = msg.MsgBody.Qr_code_info.Refund_reason
	t.Out_request_no = msg.MsgBody.Order_id
	return nil
}

type Alipay_trade_cancel struct {
	Out_trade_no string `json:"out_trade_no,omitempty"`
	Trade_no     string `json:"trade_no,omitempty"`
}

func (t Alipay_trade_cancel) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Alipay_trade_cancel 打包失败")
	}
	return string(res), nil
}

func (t Alipay_trade_cancel) GetMethod() string {
	return "alipay.trade.cancel"
}

func (t *Alipay_trade_cancel) InitRequest(tran *AlipayServices, msg *scanModel.TransMessage) gerror.IError {
	t.Trade_no = msg.MsgBody.Orig_sys_order_id
	t.Out_trade_no = msg.MsgBody.Orig_order_id
	return nil
}

type Alipay_trade_create struct {
	Out_trade_no          string        `json:"out_trade_no,omitempty"`
	Seller_id             string        `json:"seller_id,omitempty"`
	Total_amount          string        `json:"total_amount,omitempty"`
	Discountable_amount   string        `json:"discountable_amount,omitempty"`
	Undiscountable_amount string        `json:"undiscountable_amount,omitempty"`
	Buyer_logon_id        string        `json:"buyer_logon_id,omitempty"`
	Subject               string        `json:"subject,omitempty"`
	Body                  string        `json:"body,omitempty"`
	Buyer_id              string        `json:"buyer_id,omitempty"`
	Goods_detail          []goodsDetail `json:"goods_detail,omitempty"`
	Operator_id           string        `json:"operator_id,omitempty"`
	Store_id              string        `json:"store_id,omitempty"`
	Terminal_id           string        `json:"terminal_id,omitempty"`
	Extend_params         *extendParams `json:"extend_params,omitempty"`
	Timeout_express       string        `json:"timeout_express,omitempty"`
	Royalty_info          *royaltyInfo  `json:"royalty_info,omitempty"`
	Alipay_store_id       string        `json:"alipay_store_id,omitempty"`
	Sub_merchant          *subMerchant  `json:"sub_merchant,omitempty"`
	Merchant_order_no     string        `json:"merchant_order_no,omitempty"`
}

func (t Alipay_trade_create) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Alipay_trade_create 打包失败")
	}
	return string(res), nil
}

func (t Alipay_trade_create) GetMethod() string {
	return "alipay.trade.create"
}

func (t *Alipay_trade_create) InitRequest(tran *AlipayServices, msg *scanModel.TransMessage) gerror.IError {
	var gerr gerror.IError

	t.Buyer_id = msg.MsgBody.Qr_code_info.Sub_user_id
	t.Sub_merchant = &subMerchant{}
	t.Sub_merchant.Merchant_id = msg.MsgBody.Mcht_cd
	t.Out_trade_no = msg.MsgBody.Order_id
	t.Total_amount, gerr = DiveAmt(msg.MsgBody.Tran_amt)
	if gerr != nil {
		return gerr
	}
	t.Subject = msg.MsgBody.Mcht_nm
	if svrpid := tran.GetSvrPid(msg.MsgBody.InsIdCd); svrpid != "" {
		t.Extend_params = &extendParams{}
		t.Extend_params.Sys_service_provider_id = svrpid
	}
	if msg.MsgBody.Qr_code_info.Time_out != "" {
		secNum, err := strconv.Atoi(msg.MsgBody.Qr_code_info.Time_out)
		secNum = secNum / 60
		if err != nil {
			t.Timeout_express = tran.TranTimeOut
		} else if secNum == 0 {
			t.Timeout_express = "1m"
		} else {
			t.Timeout_express = fmt.Sprintf("%dm", secNum)
		}
	} else {
		t.Timeout_express = tran.OrderTimeOut
	}
	t.Store_id = msg.MsgBody.Qr_code_info.Store_id
	t.Terminal_id = msg.MsgBody.Term_id

	return nil
}

type Alipay_trade_pre_create struct {
	Out_trade_no          string        `json:"out_trade_no,omitempty"`
	Seller_id             string        `json:"seller_id,omitempty"`
	Total_amount          string        `json:"total_amount,omitempty"`
	Discountable_amount   string        `json:"discountable_amount,omitempty"`
	Undiscountable_amount string        `json:"undiscountable_amount,omitempty"`
	Buyer_logon_id        string        `json:"buyer_logon_id,omitempty"`
	Subject               string        `json:"subject,omitempty"`
	Body                  string        `json:"body,omitempty"`
	Goods_detail          []goodsDetail `json:"goods_detail,omitempty"`
	Operator_id           string        `json:"operator_id,omitempty"`
	Store_id              string        `json:"store_id,omitempty"`
	Terminal_id           string        `json:"terminal_id,omitempty"`
	Extend_params         *extendParams `json:"extend_params,omitempty"`
	Timeout_express       string        `json:"timeout_express,omitempty"`
	Royalty_info          *royaltyInfo  `json:"royalty_info,omitempty"`
	Sub_merchant          *subMerchant  `json:"sub_merchant,omitempty"`
	Alipay_store_id       string        `json:"alipay_store_id,omitempty"`
}

func (t Alipay_trade_pre_create) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Alipay_trade_create 打包失败")
	}
	return string(res), nil
}

func (t Alipay_trade_pre_create) GetMethod() string {
	return "alipay.trade.precreate"
}

func (t *Alipay_trade_pre_create) InitRequest(tran *AlipayServices, msg *scanModel.TransMessage) gerror.IError {
	var gerr gerror.IError

	t.Sub_merchant = &subMerchant{}
	t.Sub_merchant.Merchant_id = msg.MsgBody.Mcht_cd
	t.Out_trade_no = msg.MsgBody.Order_id
	t.Total_amount, gerr = DiveAmt(msg.MsgBody.Tran_amt)
	if gerr != nil {
		return gerr
	}
	t.Subject = msg.MsgBody.Mcht_nm
	if svrpid := tran.GetSvrPid(msg.MsgBody.InsIdCd); svrpid != "" {
		t.Extend_params = &extendParams{}
		t.Extend_params.Sys_service_provider_id = svrpid
	}
	if msg.MsgBody.Qr_code_info.Time_out != "" {
		secNum, err := strconv.Atoi(msg.MsgBody.Qr_code_info.Time_out)
		secNum = secNum / 60
		if err != nil {
			t.Timeout_express = tran.TranTimeOut
		} else if secNum == 0 {
			t.Timeout_express = "1m"
		} else {
			t.Timeout_express = fmt.Sprintf("%dm", secNum)
		}
	} else {
		t.Timeout_express = tran.OrderTimeOut
	}
	t.Store_id = msg.MsgBody.Qr_code_info.Store_id
	t.Terminal_id = msg.MsgBody.Term_id

	return nil
}

type Alipay_trade_close struct {
	Alipay_trade_query
	operator_id string
}

func (t Alipay_trade_close) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Alipay_trade_close 打包失败")
	}
	return string(res), nil
}

func (t Alipay_trade_close) GetMethod() string {
	return "alipay.trade.close"
}

type Alipay_boss_prod_submerchant_create struct {
	SubMerchantId string `json:"sub_merchant_id,omitempty"`
	ExternalId    string `json:"external_id,omitempty"`
	Name          string `json:"name,omitempty"`
	AliasName     string `json:"alias_name,omitempty"`
	ServicePhone  string `json:"service_phone,omitempty"`
	ContactName   string `json:"contact_name,omitempty"`
	ContactPhone  string `json:"contact_phone,omitempty"`
	ContactMobile string `json:"contact_mobile,omitempty"`
	ContactEmail  string `json:"contact_email,omitempty"`
	CategoryId    string `json:"category_id,omitempty"`
	Source        string `json:"source,omitempty"`
	Memo          string `json:"memo,omitempty"`
}

func (t Alipay_boss_prod_submerchant_create) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Alipay_boss_prod_submerchant_create 打包失败")
	}
	return string(res), nil
}
func (t Alipay_boss_prod_submerchant_create) GetMethod() string {
	return "alipay.boss.prod.submerchant.create"
}

func (t *Alipay_boss_prod_submerchant_create) InitRequest(tran *AlipayServices,
	msg *scanModel.TransMessage) gerror.IError {

	if msg == nil {
		return nil
	}
	return nil
}

type Contact_info struct {
	Name       string `jsoN:"name,omitempty"`
	Phone      string `jsoN:"phone,omitempty"`
	Mobile     string `jsoN:"mobile,omitempty"`
	Email      string `jsoN:"email,omitempty"`
	Type       string `jsoN:"type,omitempty"`
	Id_card_no string `jsoN:"id_card_no,omitempty"`
}
type Address_info struct {
	Province_code string `json:"province_code,omitempty"`
	City_code     string `json:"city_code,omitempty"`
	District_code string `json:"district_code,omitempty"`
	Address       string `json:"address,omitempty"`
	Longitude     string `json:"longitude,omitempty"`
	Latitude      string `json:"latitude,omitempty"`
	Type          string `json:"type,omitempty"`
}
type Bankcard_info struct {
	Card_no   string `json:"card_no,omitempty"`
	Card_name string `json:"card_name,omitempty"`
}
type Ant_merchant_expand_indirect_create struct {
	SubMerchantId       string           `json:"sub_merchant_id,omitempty"`
	ExternalId          string           `json:"external_id,omitempty"`
	Name                string           `json:"name,omitempty"`
	AliasName           string           `json:"alias_name,omitempty"`
	ServicePhone        string           `json:"service_phone,omitempty"`
	CategoryId          string           `json:"category_id,omitempty"`
	Source              string           `json:"source,omitempty"`
	BusinessLicense     string           `json:"business_license,omitempty"`
	BusinessLicenseType string           `json:"business_license_type,omitempty"`
	ContactInfo         []*Contact_info  `json:"contact_info,omitempty"`
	AddressInfo         []*Address_info  `json:"address_info,omitempty"`
	BankcardInfo        []*Bankcard_info `json:"bankcard_info,omitempty"`
	PayCodeInfo         string           `json:"pay_code_info,omitempty"`
	LogonId             string           `json:"logon_id,omitempty"`
	Memo                string           `json:"memo,omitempty"`
}

func (t Ant_merchant_expand_indirect_create) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Ant_merchant_expand_indirect_create 打包失败")
	}
	return string(res), nil
}
func (t Ant_merchant_expand_indirect_create) GetMethod() string {
	return "ant.merchant.expand.indirect.create"
}

func (t *Ant_merchant_expand_indirect_create) InitRequest(tran *AlipayServices,
	msg *scanModel.TransMessage) gerror.IError {

	if msg == nil {
		return nil
	}
	return nil
}

type Ant_merchant_expand_indirect_query struct {
	Sub_merchant_id string `json:"Sub_merchant_id,omitempty"`
	External_id     string `json:"external_id,omitempty"`
}

func (t Ant_merchant_expand_indirect_query) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "Ant_merchant_expand_indirect_query 打包失败")
	}
	return string(res), nil
}
func (t Ant_merchant_expand_indirect_query) GetMethod() string {
	return "ant.merchant.expand.indirect.query"
}

func (t *Ant_merchant_expand_indirect_query) InitRequest(tran *AlipayServices,
	msg *scanModel.TransMessage) gerror.IError {

	if msg == nil {
		return nil
	}
	return nil
}

type SettFileReq struct {
	BillType string `json:"bill_type,omitempty"`
	BillDate string `json:"bill_date,omitempty"`
}

func (t SettFileReq) ToString() (string, gerror.IError) {
	res, err := json.Marshal(t)
	if err != nil {
		return "", gerror.NewR(21001, err, "SettFileReq 打包失败")
	}
	return string(res), nil
}
func (t SettFileReq) GetMethod() string {
	return "alipay.data.dataservice.bill.downloadurl.query"
}

func (t *SettFileReq) InitRequest(tran *AlipayServices,
	msg *scanModel.TransMessage) gerror.IError {

	if msg == nil {
		return nil
	}
	return nil
}

type Alipay_boss_prod_submerchant_modify struct {
	Alipay_boss_prod_submerchant_create
}

func (t Alipay_boss_prod_submerchant_modify) GetMethod() string {
	return "alipay.boss.prod.submerchant.modify"
}

type Ant_merchant_expand_indirect_modify struct {
	Ant_merchant_expand_indirect_create
}

func (t Ant_merchant_expand_indirect_modify) GetMethod() string {
	return "ant.merchant.expand.indirect.modify"
}
