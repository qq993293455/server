package apis

import (
	"fmt"
	"net/http"

	"git.skyunion.net/igg-server-sdk/server-sdk-go/iggreq"
)

//AlarmOptions 报警可选参数
type AlarmOptions struct {
	//Title 标题
	Title string
	//Tag 标签
	Tag string
	//GameID 游戏ID
	GameID int64
}

//Alarm 报警接口
type Alarm struct {
	gateway *iggreq.GatewayRequest
}

const (
	alarmSend APIBusiness = iota
)

//NewAlarmServiceDefault 创建报警接口实例，默认超时10秒
func NewAlarmServiceDefault() *Alarm {
	return NewAlarmService(iggreq.ConfigInstance().GatewayConfig(), 10)
}

//NewAlarmService 创建报警接口实例
//
// apiToken 消费者token
//
// apiSecret 消费者secret
func NewAlarmService(gatewayConf iggreq.GatewayConfiguration, timeoutSecond int) *Alarm {
	return &Alarm{
		gateway: iggreq.NewGatewayRequest(gatewayConf, timeoutSecond),
	}
}

//Send 发送告警
//
// token 报警点
//
// content 报警内容
//
// opts 可选项
func (a *Alarm) Send(token, content string, opts *AlarmOptions) IGGError {
	// 2020.6.3 王亮：我刚才问了郁菲，他报警和推送的原来的签名都不用了
	//query := iggreq.NewQueryMaker(AlarmSecCode)
	query := iggreq.NewQueryMaker("")
	query.AddParam("token", token)
	query.AddParam("content", content)

	if opts != nil {
		if opts.GameID != -1 {
			query.AddParam("game_id", opts.GameID)
		}
		if opts.Tag != "" {
			query.AddParam("tag", opts.Tag)
		}
		if opts.Title != "" {
			query.AddParam("title", opts.Title)
		}
	}
	s, e, h := a.gateway.Post(iggreq.AlarmAPIPath, query)
	if !e.IsSuccess() {
		return alarmInternalErrorToSDKExceptton(alarmSend, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return alarmInternalErrorToSDKExceptton(alarmSend, e, h)
}

func alarmInternalErrorToSDKExceptton(bus APIBusiness, e *iggreq.InternalError, hdr http.Header) IGGError {
	code := IGGError{}
	if !e.IsSuccess() {
		code = NewIGGError(ALARM, bus, int(e.Type))
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
