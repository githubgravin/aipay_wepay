package main

import (
	"fmt"
	"os"
	"unGateWay/Config"
	"unGateWay/HttpsSvr"
)

var UNCFGFILE string = "./etc/unGateCfg.ini"

func main() {
	//加载配置模块
	gerr := Config.InitGlobalCfg(UNCFGFILE)
	if gerr != nil {
		fmt.Fprintln(os.Stderr, gerr)
		return
	}

	//DB环境变量
	os.Setenv("NLS_LANG", "AMERICAN_AMERICA.AL32UTF8")
	//模块加载完成
	err := HttpsSvr.InitModule("HTTPSSVR")
	if err != nil {
		fmt.Fprintln(os.Stderr, "[HTTPSSVR]加载失败;", err)
		return
	}

	fmt.Println("程序启动成功")

	//select {}
}
