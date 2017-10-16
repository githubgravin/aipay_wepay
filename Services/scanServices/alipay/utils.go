package alipay

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/axgle/mahonia"
	"golib/gerror"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"
)

func RsaSignBase64Url(privSign *rsa.PrivateKey, postUrl url.Values) (string, gerror.IError) {

	sortKeys := make([]string, 0)
	for key := range postUrl {
		sortKeys = append(sortKeys, key)
	}
	sort.Strings(sortKeys)

	var tmpBuf string
	var signBuf string
	for key := range sortKeys {
		tmpBuf = fmt.Sprintf("%s=%s&", sortKeys[key], postUrl.Get(sortKeys[key]))
		signBuf += tmpBuf
	}
	signBuf = signBuf[:len(signBuf)-1]

	ciperdata, err := RsaSignSha1(privSign, []byte(signBuf))
	if err != nil {
		return "", gerror.NewR(1011, err, "RsaSignSha1 error")
	}
	return base64.StdEncoding.EncodeToString(ciperdata), nil
}

func RsaSignSha1Base64(privSign *rsa.PrivateKey, data []byte) (string, gerror.IError) {
	h := sha1.New()
	h.Write([]byte(data))
	digest := h.Sum(nil)

	ciphertext, err := rsa.SignPKCS1v15(nil, privSign, crypto.SHA1, digest)
	if err != nil {
		return "", gerror.NewR(1012, err, "rsa.SignPKCS1v15 error;")
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func RsaSignSha1(privSign *rsa.PrivateKey, data []byte) ([]byte, gerror.IError) {
	h := sha1.New()
	h.Write([]byte(data))
	digest := h.Sum(nil)

	ciphertext, err := rsa.SignPKCS1v15(nil, privSign, crypto.SHA1, digest)
	if err != nil {
		return nil, gerror.NewR(1012, err, "rsa.SignPKCS1v15 error;")
	}
	return ciphertext, nil
}

//金额除100
func DiveAmt(srcAmt string) (string, gerror.IError) {
	amt, err := strconv.Atoi(srcAmt)
	if err != nil {
		return "", gerror.NewR(1013, err, "转换金额失败")
	}
	damt := float64(amt) / 100.00
	return fmt.Sprintf("%.2f", damt), nil
}

func RespConv(srcRet, subRet string) string {
	switch srcRet {
	case ALI_RESP_SUCCESS:
		return "00"
	case ALI_RESP_TRANING:
		return "W3"
	default:
		if RespConv, ok := ALI_ERR_CD_CONV[subRet]; ok {
			return RespConv
		} else {
			return "96"
		}
	}
	return "96"
}

func TradeStatConv(srcSt, subRet string) string {
	switch srcSt {
	case "WAIT_BUYER_PAY":
		return "98"
	case "TRADE_CLOSED":
		return "S5"
	case "TRADE_SUCCESS":
		return "00"
	case "TRADE_FINISHED":
		return "00"
	default:
		return "98"
	}
	return "96"
}

/*UTF8 转码到 GBK*/
func UTF8tGBK(src []byte) (dst []byte) {
	u2g := mahonia.NewEncoder("GBK")
	dst = []byte(u2g.ConvertString(string(src)))
	return dst
}

/*GBK 转码到 UTF8*/
func GBKtUTF8(src string) (dst string) {
	u2g := mahonia.NewDecoder("GBK")
	tmp := u2g.ConvertString(src)
	return tmp
}

/*时间格式转化*/
func TimeConv(timeStr string) string {
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return ""
	}
	nt := t.Format("20060102150405")
	return nt

}
func Yuan2Points(amtPoints string) string {
	i, err := strconv.ParseFloat(amtPoints, 2)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%.0f", i*100)
}

func DownLoadFile(downDir, reqUrl string) (string, gerror.IError) {
	pUrl, err := url.Parse(reqUrl)
	if err != nil {
		fmt.Println(err)
		return "", gerror.NewR(30030, err, "reqUrl[%s] 解析失败;", reqUrl)
	}
	uVal := pUrl.Query()
	/*文件名*/
	fileName := uVal.Get("downloadFileName")
	/*账务日期*/
	settDate := uVal.Get("bizDates")
	dir := downDir + "/" + settDate
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", gerror.NewR(30030, err, "本地创建目录失败[%s]", dir)
	}

	localFile := dir + "/" + fileName
	fp, err := os.OpenFile(localFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return " ", gerror.NewR(30030, err, "本地创建文件失败[%s]", localFile)
	}
	defer fp.Close()

	req, _ := http.NewRequest("GET", reqUrl, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", gerror.NewR(20020, err, "GET: Client.Do error; ")
	}
	defer resp.Body.Close()

	_, err = io.Copy(fp, resp.Body)
	if err != nil {
		return "", gerror.NewR(20020, err, "GET: Client.Do error; ")
	}
	return localFile, nil
}
