/**
* @Author: cl
* @Date: 2021/4/13 19:24
 */
package httpsserver

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	conf "httpsproxy/config"
	"io/ioutil"
	"mq.code.sangfor.org/SaaS/gobase/mlog"
	"net/http"
	"net/url"
	"strings"
)

func TransmitServer(listenAdress string) {
	url := "/"
	mlog.Infof("url:[%s]", url)
	http.HandleFunc(url, postMethod)
	err := http.ListenAndServeTLS(listenAdress, conf.Cfg.Proxy.ServerCrt, conf.Cfg.Proxy.ServerKey, nil)
	if err != nil {
		mlog.Fatalf("err:%v", err)
	}
}

type Resp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func postMethod(w http.ResponseWriter, r *http.Request) {
	var tr *http.Transport
	if conf.Cfg.Proxy.PHost == "" || conf.Cfg.Proxy.PPort == 0 {
		mlog.Warn("direct send!")
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}else {
		proxyAddr := fmt.Sprintf("http://%s:%d", conf.Cfg.Proxy.PHost, conf.Cfg.Proxy.PPort)
		mlog.Debugf("proxyAddr:%s", proxyAddr)
		u, err := url.Parse(proxyAddr)
		if err != nil {
			mlog.Warnf("err:%v, direct send!", err)
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		} else {
			tr = &http.Transport{
				Proxy:           http.ProxyURL(u),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}
	}
	client := &http.Client{Transport: tr}
	url := fmt.Sprintf("https://%s:%d%s", conf.Cfg.Proxy.THost, conf.Cfg.Proxy.TPort, r.URL.Path)
	mlog.Debugf("url:%s", url)
	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		b, _ := json.Marshal(Resp{Code: 10002, Msg: fmt.Sprintf("url [%s] NewRequest is failed", url)})
		_, _ = w.Write(b)
		return
	}
	for k, vv := range r.Header {
		for _, v := range vv {
			mlog.Infof("k:%v, v:%v", k, v)
			req.Header.Set(k, v)
		}
	}
	addr := strings.Split(r.RemoteAddr, ":")[0]
	xforwarded := req.Header.Get("X-Forwarded-For")
	if xforwarded == "" {
		req.Header.Set("X-Forwarded-For", addr)
	} else {
		req.Header.Add("X-Forwarded-For", addr)
	}

	resp, err := client.Do(req)
	if err != nil {
		b, _ := json.Marshal(Resp{Code: 10003, Msg: fmt.Sprintf("client [%s] Do is failed", url)})
		_, _ = w.Write(b)
		return
	}
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mlog.Warnf("err:%v", err)
		b, _ := json.Marshal(Resp{Code: 10004, Msg: "resp ReadAll is failed"})
		_, _ = w.Write(b)
		return
	}
	_, _ = w.Write(rbody)
	return
}