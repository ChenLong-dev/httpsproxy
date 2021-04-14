/**
* @Author: cl
* @Date: 2021/4/13 19:27
 */
package proxy


import (
	conf "httpsproxy/config"
	"io"
	"mq.code.sangfor.org/SaaS/gobase/mlog"
	"net"
	"net/http"
	"strings"
	"time"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleHttps(w, r)
		mlog.Infof("[HTTPS] w:%v", w)
		mlog.Infof("[HTTPS] r:%v", r)
	} else {
		handleHttp(w, r)
		mlog.Infof("[HTTP] w:%v", w)
		mlog.Infof("[HTTP] r:%v", r)
	}
}

func handleHttps(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 60*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		mlog.Fatalf("err:%v", err)
		return
	}
	mlog.Infof("==== [type:%s] host:%s,  method:%s, header:%v, host:%v", conf.Cfg.Proxy.Type, r.Host, r.Method, r.Header, r.RemoteAddr)
	addr := strings.Split(r.RemoteAddr, ":")[0]
	r.Header.Add("X-Forwarded-For", addr)
	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		mlog.Fatalf("err:%v", ok)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		mlog.Fatalf("err:%v", err)
	}
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		mlog.Fatalf("err:%v", err)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			mlog.Infof("k:%v, v:%v", k, v)
			dst.Add(k, v)
		}
	}
}