package iggsdk

import (
	"coin-server/common/errmsg"
	"git.skyunion.net/igg-server-sdk/server-sdk-go/apis"
)

type ApiIpI interface {
	GetIp(ip string) ([]apis.IPInfo, *errmsg.ErrMsg)
}

type apiIp struct {
	ai *apis.IPQuery
}

var ipIns *apiIp

func InitIpIns() {
	ipIns = &apiIp{ai: apis.NewIPQueryServiceDefault()}
}

func GetIpIns() ApiIpI {
	return ipIns
}

func (ai *apiIp) GetIp(ip string) ([]apis.IPInfo, *errmsg.ErrMsg) {
	info, err := ai.ai.GetIP(ip)
	if !err.IsSuccess() {
		return nil, errmsg.NewInternalErr(err.Dump())
	}
	return info, nil
}
