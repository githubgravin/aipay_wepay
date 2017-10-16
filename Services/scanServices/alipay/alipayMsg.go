package alipay

import (
	"golib/gerror"
	"unGateWay/Services/scanServices/scanModel"
)

//域顺序按字母排序  签名要求
type CommonRequest struct {
	App_auth_token string `json:"app_auth_token,omitempty"`
	App_id         string `json:"app_id,omitempty"`
	Biz_content    string `json:"biz_content,omitempty"`
	Charset        string `json:"charset,omitempty"`
	Format         string `json:"format,omitempty"`
	Method         string `json:"method,omitempty"`
	Notify_url     string `json:"notify_url,omitempty"`
	Sign_type      string `json:"sign_type,omitempty"`
	Sign           string `json:"sign,omitempty"`
	Timestamp      string `json:"timestamp,omitempty"`
	Version        string `json:"version,omitempty"`
}

type IBizCon interface {
	ToString() (string, gerror.IError)
	GetMethod() string
}

func (t *CommonRequest) InitBase() {
	t.Format = "JSON"
	t.Charset = "GBK"
	t.Sign_type = "RSA"
	t.Version = "1.0"
}

type IAlipayReq interface {
	IBizCon
	InitRequest(tran *AlipayServices, msg *scanModel.TransMessage) gerror.IError
	//Clone() IAlipayReq
}

type IAlipayRsp interface {
	LoadResponse(rspMsg string, msg *scanModel.TransMessage) gerror.IError
	//Clone() IAlipayRsp
}
