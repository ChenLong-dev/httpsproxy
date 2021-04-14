/**
* @Author: cl
* @Date: 2021/4/13 17:25
 */
package main


import (
	"fmt"
	conf "httpsproxy/config"
	"httpsproxy/httpsserver"
	"mq.code.sangfor.org/SaaS/gobase/config"
	"mq.code.sangfor.org/SaaS/gobase/mlog"
	"net"
)

func Init() {
	// 配置文件初始化
	config.InitExt(&conf.Cfg)
	fmt.Println("server start mode: ", conf.Cfg.AppMode)

	// 日志初始化
	logCfg := conf.Cfg.Log
	_ = mlog.Init(&mlog.Params{
		Path:       logCfg.Path,
		MaxSize:    logCfg.MaxSize,
		MaxBackups: logCfg.MaxBackups,
		MaxAge:     logCfg.MaxAge,
		Level:      logCfg.Level,
	})
}

func main() {
	Init()
	listenAdress := fmt.Sprintf("%s:%d", conf.Cfg.Proxy.Host, conf.Cfg.Proxy.Port)
	if !checkAdress(listenAdress) {
		mlog.Fatal("-L listen address format incorrect.Please check it")
	}

	mlog.Infof("====== [type:%s] listenAdress:%s ======", conf.Cfg.Proxy.Type, listenAdress)
	if conf.Cfg.Proxy.Type == TransmitServer {
		mlog.Infof("=== proxy addr [%s:%d], target addr [%s:%d] ===", conf.Cfg.Proxy.PHost, conf.Cfg.Proxy.PPort,
			conf.Cfg.Proxy.THost, conf.Cfg.Proxy.TPort)
		httpsserver.TransmitServer(listenAdress)
	} else if conf.Cfg.Proxy.Type == ProxyAgent {
		mlog.Infof("=== proxy addr [%s:%d], target addr [%s:%d] ===", conf.Cfg.Proxy.PHost, conf.Cfg.Proxy.PPort)
		httpsserver.ProxyAgent(listenAdress)
	} else if conf.Cfg.Proxy.Type == ProxyServer {
		httpsserver.ProxyServe(listenAdress)
	} else {
		mlog.Fatalf("type (%s) is error", conf.Cfg.Proxy.Type)
		return
	}
}

func checkAdress(adress string) bool {
	_, err := net.ResolveTCPAddr("tcp", adress)
	if err != nil {
		return false
	}
	return true

}