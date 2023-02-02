package self_ip

import (
	"fmt"
	"net"
	"strings"
	"time"

	"coin-server/common/consulkv"
	"coin-server/common/values/env"

	"github.com/guonaihong/gout"
)

var SelfIpLan string
var SelfIPWan string

type Resp struct {
	IPInfo IPInfo `json:"ip_info"`
}
type IPInfo struct {
	IP       string   `json:"ip"`       //ip地址
	Country  string   `json:"country"`  //国家
	Province string   `json:"province"` //省份
	City     string   `json:"city"`     //城市
	Location struct { //经纬度
		AccuracyRadius uint16  `json:"accuracy_radius"`
		Latitude       float64 `json:"latitude"`
		Longitude      float64 `json:"longitude"`
		MetroCode      uint    `json:"metro_code"`
		TimeZone       string  `json:"time_zone"`
	} `json:"location"`
}

type IpServiceAddrS struct {
	Lan string `json:"lan"`
	Wan string `json:"wan"`
}

func Init(cnf *consulkv.Config) {
	isa := &IpServiceAddrS{}
	err := cnf.Unmarshal("ip-service", isa)
	if err != nil {
		panic(err)
	}
	SelfIpLan, SelfIPWan = Self(isa)
}

func InitLan(cnf *consulkv.Config) {
	isa := &IpServiceAddrS{}
	err := cnf.Unmarshal("ip-service", isa)
	if err != nil {
		panic(err)
	}
	SelfIpLan = SelfLan(isa)
}

func Self(isa *IpServiceAddrS) (lan, wan string) {
	isa.Lan = strings.TrimSpace(isa.Lan)
	isa.Wan = strings.TrimSpace(isa.Wan)
	if isa.Lan == "" && isa.Wan == "" {
		panic("IpServiceAddrS is empty")
	}
	lan = getIp(isa.Lan)
	wan = getIp(isa.Wan)
	if lan == "" && wan == "" {
		panic(fmt.Sprintf("can not got ip from %s and %s", isa.Wan, isa.Lan))
	}
	if lan == "" {
		lan = wan
	}
	if wan == "" {
		wan = lan
	}
	return
}

func SelfLan(isa *IpServiceAddrS) string {
	isa.Lan = strings.TrimSpace(isa.Lan)
	if isa.Lan == "" {
		panic("IpServiceAddrS Lan is empty")
	}
	lan := getIp(isa.Lan)

	if lan == "" {
		panic(fmt.Sprintf("can not got ip from %s ", isa.Lan))
	}
	return lan
}

func getIp(addr string) string {
	if addr == "" {
		return ""
	}
	ipInfo := &Resp{}
	err := gout.GET(fmt.Sprintf("http://%s/?language=cn&format=string", addr)).SetTimeout(time.Second * 3).BindJSON(ipInfo).Do()
	if err != nil {
		panic(err)
	}
	return ipInfo.IPInfo.IP
}

func GetTCPAddrS() (lan, wan string) {
	p := env.GetTCPPort()
	return net.JoinHostPort(SelfIpLan, p), net.JoinHostPort(SelfIPWan, p)
}

func GetHTTPAddrS() (lan, wan string) {
	p := env.GetHTTPPort()
	return net.JoinHostPort(SelfIpLan, p), net.JoinHostPort(SelfIPWan, p)
}

func GetPPROFAddrS() (lan, wan string) {
	p := env.GetPPROFPort()
	return net.JoinHostPort(SelfIpLan, p), net.JoinHostPort(SelfIPWan, p)
}
