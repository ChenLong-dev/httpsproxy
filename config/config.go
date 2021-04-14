/**
 .oooooo..o            .o.                  .o.             .oooooo..o
d8P'    `Y8           .888.                .888.           d8P'    `Y8
Y88bo.               .8"888.              .8"888.          Y88bo.
 `"Y8888o.          .8' `888.            .8' `888.          `"Y8888o.
     `"Y88b        .88ooo8888.          .88ooo8888.             `"Y88b
oo     .d8P       .8'     `888.        .8'     `888.       oo     .d8P
8""88888P'       o88o     o8888o      o88o     o8888o      8""88888P'

* @Author: cl
* @Date: 2021/3/9 21:16
*/
package config

var Cfg ConfigExt

//属性设计为指针类型
type ConfigExt struct {
	AppMode string
	Mdbg    *MdbgCfg
	Proxy   *Proxy
	Log     *Log
}

type MdbgCfg struct {
	Host   string
	Port   int
	Enable int
}

type Proxy struct {
	Host      string
	Port      int
	ServerCrt string
	ServerKey string
	PHost     string
	PPort     int
	THost     string
	TPort     int
	Type      string
}

type Log struct {
	Path       string
	Level      int
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

