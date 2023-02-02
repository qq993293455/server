package apis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"git.skyunion.net/igg-server-sdk/server-sdk-go/iggreq"
)

// IPQuery ip 查询
type IPQuery struct {
	gateway *iggreq.GatewayRequest
}

// IPInfo IP 查询结果值
type IPInfo struct {
	City          string
	Continent     string
	ContinentCode string
	Country       string
	CountryCode   string
	IP            string
	ISP           string
	Region        string
	Subregion     string
}

const (
	ipqGetIP APIBusiness = iota
)

//NewIPQueryServiceDefault 新建IP查询实例(默认超时10秒)
func NewIPQueryServiceDefault() *IPQuery {
	return NewIPQueryService(iggreq.ConfigInstance().GatewayConfig(), 10)
}

//NewIPQueryService 新建IP查询实例
//
// apiToken 消费者token
//
// apiSecret 消费者secret
func NewIPQueryService(gatewayConf iggreq.GatewayConfiguration, timeoutSecond int) *IPQuery {
	return &IPQuery{
		gateway: iggreq.NewGatewayRequest(gatewayConf, timeoutSecond),
	}
}

//GetIP 获取IP地址信息
//
// iplist 以`|`分割的的IP地址列表
func (a *IPQuery) GetIP(iplist string) ([]IPInfo, IGGError) {
	query := iggreq.NewQueryMaker("")
	query.AddParam("ip", iplist)

	res := make([]IPInfo, 0, len(strings.Split(iplist, "|")))
	s, e, h := a.gateway.Get(iggreq.IPQueryAPIPath, query)
	if !e.IsSuccess() {
		return res, iqpInternalErrorToSDKExceptton(ipqGetIP, e, h)
	}

	traceid := e.TraceID
	d, e := iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	switch o := d.(type) {
	case map[string]interface{}:
		buf, _ := json.Marshal(o)
		if _, ok := o["City"]; ok {
			// 单个结果
			var info IPInfo
			json.Unmarshal(buf, &info)
			res = append(res, info)
		} else {
			// 多个结果
			var infoMap map[string]IPInfo
			json.Unmarshal(buf, &infoMap)
			for _, info := range infoMap {
				res = append(res, info)
			}
		}
	default:
	}
	return res, iqpInternalErrorToSDKExceptton(ipqGetIP, e, h)
}

func iqpInternalErrorToSDKExceptton(bus APIBusiness, e *iggreq.InternalError, hdr http.Header) IGGError {
	code := IGGError{}
	if !e.IsSuccess() {
		code = NewIGGError(IPQUERY, bus, int(e.Type))
		code.underlyingCode = fmt.Sprint(e.Code)
		code.extensionCode = e.Msg
	}
	if e.TraceID != "" {
		code.traceIDs = append(code.traceIDs, e.TraceID)
	}
	if hdr != nil {
		if values, ok := hdr["X-IGG-TRACEID"]; ok {
			code.traceIDs = append(code.traceIDs, values[0])
		}
	}
	return code
}
