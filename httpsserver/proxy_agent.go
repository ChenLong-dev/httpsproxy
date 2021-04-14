/**
* @Author: cl
* @Date: 2021/4/13 19:22
 */
package httpsserver

import (
	"bufio"
	"bytes"
	"fmt"
	conf "httpsproxy/config"
	"io"
	"mq.code.sangfor.org/SaaS/gobase/mlog"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func ProxyAgent(listenAdress string) {
	conn, err := net.Listen("tcp", listenAdress)
	if err != nil {
		mlog.Warnf("err:%v", err)
		return
	}

	for {
		client, err := conn.Accept()
		if err != nil {
			mlog.Warnf("err:%v", err)
			return
		}
		go HandleAgent(client)
	}
}

func HandleAgent(client net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			mlog.Warnf("err:%v", err)
		}
	}()
	if client == nil {
		mlog.Warn("client == nil")
		return
	}
	remoteAddr := client.RemoteAddr().String()
	mlog.Infof("client tcp tunnel connection: %s -> %s", remoteAddr, client.LocalAddr().String())
	defer client.Close()

	var b [5120]byte
	n, err := client.Read(b[:]) //读取应用层的所有数据
	if err != nil || bytes.IndexByte(b[:], '\n') == -1 {
		mlog.Warnf("err:%v", err)
		return
	}

	var method, host, address string
	mlog.Debug(string(b[:n]))
	_, _ = fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	mlog.Infof("method:%s, host:%s", method, host)

	readdr := strings.Split(remoteAddr, ":")[0]
	remoteAddrs, err := getForwordFromBit(string(b[:n]))
	if err == nil && remoteAddrs != "" {
		remoteAddrs = fmt.Sprintf("%s,%s", remoteAddrs, readdr)
	} else {
		remoteAddrs = readdr
	}

	mlog.Debugf("remoteAddr:%s, remoteAddrs:%s", remoteAddr, remoteAddrs)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		mlog.Warnf("url parse is failed, host:%s", host)
		address = host
	} else {
		if hostPortURL.Opaque == "443" || hostPortURL.Opaque == "8080" {
			address = fmt.Sprintf("%s:%s", hostPortURL.Scheme, hostPortURL.Opaque)
		} else {
			if strings.Index(hostPortURL.Host, ":") == -1 {
				address = fmt.Sprintf("%s:80", hostPortURL.Host)
			} else {
				address = hostPortURL.Host
			}
		}
	}

	server, err := proxyDial("tcp", address, remoteAddrs)
	if err != nil {
		mlog.Warnf("err:%v", err)
		return
	}
	defer server.Close()
	mlog.Infof("server tcp tunnel connection: %s -> %s", server.LocalAddr().String(), server.RemoteAddr().String())
	if method == http.MethodConnect {
		mlog.Infof("method:%s success!", method)
		_, _ = fmt.Fprint(client, "HTTP/1.1 200 Connection established xxxx cl \r\n\r\n")
	} else {
		mlog.Debugf("server write to, method:%s", method)
		_, _ = server.Write(b[:n])
	}

	//进行转发
	go func() {
		io.Copy(server, client)
	}()
	io.Copy(client, server) //阻塞转发
}

//Dial 建立一个传输通道
func proxyDial(network, addr, remote string) (net.Conn, error) {
	proxyAddr := fmt.Sprintf("http://%s:%d", conf.Cfg.Proxy.PHost, conf.Cfg.Proxy.PPort)
	//建立到代理服务器的传输层通道
	var c net.Conn
	c, err := func() (net.Conn, error) {
		u, _ := url.Parse(proxyAddr)
		mlog.Infof("proxy address:%s", u.Host)
		// Dial and create client connection.
		c, err := net.DialTimeout("tcp", u.Host, time.Second*5)
		if err != nil {
			mlog.Warnf("err:%v", err)
			return nil, err
		}
		reqURL, err := url.Parse("http://" + addr)
		if err != nil {
			mlog.Warnf("err:%v", err)
			return nil, err
		}
		req, err := http.NewRequest(http.MethodConnect, reqURL.String(), nil)
		if err != nil {
			mlog.Warnf("err:%v", err)
			return nil, err
		}

		req.Close = false
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.3")
		req.Header.Set("X-Forwarded-For", remote)
		err = req.Write(c)
		if err != nil {
			mlog.Warnf("err:%v", err)
			return nil, err
		}
		mlog.Debugf("addr:%s, Header:%v, req:%v", addr, req.Header, req)

		resp, err := http.ReadResponse(bufio.NewReader(c), req)
		if err != nil {
			mlog.Warnf("err:%v", err)
			return nil, err
		}
		//defer resp.Body.Close()

		mlog.Debug(resp.StatusCode, resp.Status, resp.Proto, resp.Header)
		if resp.StatusCode != 200 {
			err = fmt.Errorf("Connect server using proxy error, StatusCode [%d]\n", resp.StatusCode)
			mlog.Warnf("err:%v", err)
			return nil, err
		}
		return c, err
	}()
	if c == nil || err != nil { //代理异常
		mlog.Warnf("proxy exception,conn:%v, err:%v, local direct transmit, network:%s, addr:%s", c, err, network, addr)
		return net.Dial(network, addr)
	}
	mlog.Infof("proxy is success, tunnel info: %s -> %s", c.LocalAddr().String(), c.RemoteAddr().String())
	return c, err
}

func getForwordFromBit(re string) (addr string, err error) {
	/*
		CONNECT 10.107.66.86:8081 HTTP/1.1
		Host: 10.107.66.86:8081
		User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.3
		X-Forwarded-For: 10.107.63.170
	*/
	pos := strings.Index(re, "X-Forwarded-For")
	if pos < 0 {
		err = fmt.Errorf("not found X-Forwarded-For from re")
		mlog.Warnf("not found X-Forwarded-For from [%s]", re)
		return
	}
	mlog.Debugf("pos:%d", pos)
	xforwarded := re[pos:]
	addrs := strings.Split(xforwarded, ":")
	if len(addrs) != 2 {
		err = fmt.Errorf("split xforwarded (len:%d), xforwarded:%s", len(addrs), xforwarded)
		mlog.Warnf("err:%v", err)
		return
	}
	return compressStr(addrs[1]), nil
}

//利用正则表达式压缩字符串，去除空格或制表符
func compressStr(str string) string {
	if str == "" {
		return ""
	}
	//匹配一个或多个空白符的正则表达式
	reg := regexp.MustCompile("\\s+")
	return reg.ReplaceAllString(str, "")
}