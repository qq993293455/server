package apis

// 内部消息网关,部分实现,
// 参见 https://git.skyunion.net/message/message.skyunion.net/wikis/内部消息网关接入文档

import (
	"encoding/json"
	"fmt"
	"net/http"

	"git.skyunion.net/igg-server-sdk/server-sdk-go/iggreq"
)

//MessageBaseParams 消息基本属性结构
type MessageBaseParams struct {
	// id 响应的唯一消息ID
	ID string `json:"id"`
	// Title 消息的标题，有些消息不支持标题，设置成空字符串即可
	Title string `json:"title"`
	// Body  消息的正文，有些消息不支持正文，设置成空字符串即可
	Body string `json:"body"`
	// Tag 消息的标签，用于给消息分组，请务必要设置该字段
	Tag string `json:"tag"`
	// Timestamp 消息时间戳，精确到秒
	Timestamp int64 `json:"timestamp"`
	// Async 是否异步消息，异步消息能减少服务端的响应时间
	Async bool `json:"async"`
	// Delay 消息延迟发送的时间，一旦设置了大于0的延迟，消息会被自动设置为异步消息
	Delay uint64 `json:"delay"`
	// Priority 消息的优先级，0-100，数字越大优先级越低，同步消息为0，异步消息默认为30
	Priority int `json:"priority"`
}

//IGGReceiver IGG用户结构
type IGGReceiver struct {
	//账号，可以是内网用户ID(uint64)、外网用户ID(uint64)或是用户名(string)
	Account interface{} `json:"account"`
	//账号类型，1：用户名，2：内网用户ID，3：外网用户ID
	Type int `json:"type"`
}

//MessageGateway 消息网关接口
type MessageGateway struct {
	gateway  *iggreq.GatewayRequest
	apiToken string
}

const (
	gatewaySendSingle APIBusiness = iota
	gatewaySendMultip
	gatewaySendRoom
	gatewaySendMyIGG
	gatewaySendMail
	gatewayWXEntPerson
	gatewayWXEntDepartment
	gatewayWXEntCompany
	gatewayIGGSMS
	gatewayIGGSMSTempl
	gatewayIGGTTSTempl
	gatewayIGGSupport
)

//NewMessageGateDefault 创建消息网关接口实例(默认超时10秒)
func NewMessageGateDefault() *MessageGateway {
	return NewMessageGate(iggreq.ConfigInstance().GatewayConfig(), 10)
}

//NewMessageGate 创建消息网关接口实例
//
// apiToken 消费者token
//
// apiSecret 消费者secret
func NewMessageGate(gatewayConf iggreq.GatewayConfiguration, timeoutSeconds int) *MessageGateway {
	return &MessageGateway{
		gateway:  iggreq.NewGatewayRequest(gatewayConf, timeoutSeconds),
		apiToken: gatewayConf.GatewayToken,
	}
}

//SendMsgSingle 向个人发送消息
//
// baseParams 消息基本属性
//
// reciver 接收者
//
// url 消息携带URL，不同消息对于链接的显示不太一样
func (a *MessageGateway) SendMsgSingle(baseParams MessageBaseParams, reciver IGGReceiver, url string) IGGError {
	query := &msgSingle{
		Code:              "0000",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		URL:               url,
		Receiver:          reciver,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewaySendSingle, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewaySendSingle, e, h)
}

//SendMsgMultiple 向多人发送消息
//
// baseParams 消息基本属性
//
// reciver 接收者
//
// url 消息携带URL，不同消息对于链接的显示不太一样
//
// source 消息来源频道，1：公告，2：EIP，3：活动，4：任务系统，5：邮件，6：报警系统，7：其它
func (a *MessageGateway) SendMsgMultiple(baseParams MessageBaseParams, reciver []IGGReceiver, url string, source int) IGGError {
	query := &msgMultiple{
		Code:              "0001",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		URL:               url,
		Receivers:         reciver,
		Source:            source,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewaySendMultip, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewaySendMultip, e, h)
}

//SendMsgRoom 向群组发送消息，发送到IC群组
//
// baseParams 消息基本属性
//
// room 接收者
//
// at 艾特某人，可以设置多个，逗号隔开
//
// ty 内容类型，1：文本，2：图片
func (a *MessageGateway) SendMsgRoom(baseParams MessageBaseParams, room uint64, at string, ty int) IGGError {
	query := &msgRoom{
		Code:              "0002",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		Room:              room,
		At:                at,
		Type:              ty,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewaySendRoom, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewaySendRoom, e, h)
}

//SendMsgMyIGG MyIGG消息
//
// baseParams 消息基本属性
//
// reciver 接收者
//
// url 消息携带URL，不同消息对于链接的显示不太一样
//
// ty 内容类型，1：文本，2：图片
//
// source 消息来源频道，"ALARM"：警告，"EIP"：EIP，"NOTICE"：通知
func (a *MessageGateway) SendMsgMyIGG(baseParams MessageBaseParams, reciver IGGReceiver, url string, ty int, source string) IGGError {
	query := &msgMyIGG{
		Code:              "0003",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		Receiver:          reciver,
		URL:               url,
		Type:              ty,
		Source:            source,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewaySendMyIGG, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewaySendMyIGG, e, h)
}

//SendMsgMail 邮件消息，发送一封IGG邮件，不能设置发件人，收件人、抄送、密送最多不超过50个
//
// baseParams 消息基本属性
//
// reciver 接收者
//
// cc 抄送收件人
//
// bcc 密送收件人
//
// isHTML
func (a *MessageGateway) SendMsgMail(baseParams MessageBaseParams, reciver, cc, bcc []IGGReceiver, isHTML bool) IGGError {
	query := &msgMail{
		Code:              "0004",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		Receivers:         reciver,
		Ccs:               cc,
		Bccs:              bcc,
		HTML:              isHTML,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewaySendMail, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewaySendMail, e, h)
}

//SendMsgWXEntPerson 企业微信个人消息（消息代码：0005），发送给个人
//
// baseParams 消息基本属性
//
// reciver 接收者
//
// source 消息来源频道，0：登录认证提醒，3：报警&错误通知，4：提醒和通知，6：公司公告(Notice)，1000002：IGG员工服务中心(China)
func (a *MessageGateway) SendMsgWXEntPerson(baseParams MessageBaseParams, reciver IGGReceiver, source string) IGGError {
	query := &msgWXEntPerson{
		Code:              "0005",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		Receiver:          reciver,
		Source:            source,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewayWXEntPerson, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewayWXEntPerson, e, h)
}

//SendMsgWXEntDepartment 企业微信群发消息(部门)（消息代码：0006），企业微信部门群发消息，部门群发，不要随便发
//
// baseParams 消息基本属性
//
// departID 接收部门
//
// source 消息来源频道，0：登录认证提醒，3：报警&错误通知，4：提醒和通知，6：公司公告(Notice)，1000002：IGG员工服务中心(China)
func (a *MessageGateway) SendMsgWXEntDepartment(baseParams MessageBaseParams, departID uint64, source string) IGGError {
	query := &msgWXEntDepartment{
		Code:              "0006",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		DepartmentID:      departID,
		Source:            source,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewayWXEntDepartment, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewayWXEntDepartment, e, h)
}

//SendMsgWXEntCompany 企业微信群发消息(公司)（消息代码：0007），企业微信部门群发消息，部门群发，不要随便发
//
// baseParams 消息基本属性
//
// companyID 公司ID
//
// source 消息来源频道，0：登录认证提醒，3：报警&错误通知，4：提醒和通知，6：公司公告(Notice)，1000002：IGG员工服务中心(China)
func (a *MessageGateway) SendMsgWXEntCompany(baseParams MessageBaseParams, companyID uint64, source string) IGGError {
	query := &msgWXEntCompany{
		Code:              "0007",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		CompanyID:         companyID,
		Source:            source,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewayWXEntCompany, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewayWXEntCompany, e, h)
}

//SendMsgIggSMS IGG短信消息 （消息代码：0008），不支持标题，可以自定义正文，但是短信会带上IGG签名
//
// baseParams 消息基本属性
//
// companyID 公司ID
func (a *MessageGateway) SendMsgIggSMS(baseParams MessageBaseParams, receiver IGGReceiver) IGGError {
	query := &msgIggSMS{
		Code:              "0008",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		Receiver:          receiver,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewayIGGSMS, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewayIGGSMS, e, h)
}

//SendMsgIggSMSTemplate IGG短信模板消息 （消息代码：0009），比较特殊，不能设置自定义标题和正文，只支持固定模版发送
//
// baseParams 消息基本属性
//
// receiver 收件人
//
// templatekey 短信的模版KEY
//
// templateParams
func (a *MessageGateway) SendMsgIggSMSTemplate(baseParams MessageBaseParams, receiver IGGReceiver, templatekey string, templateParams interface{}) IGGError {
	query := &msgIggSMSTemplate{
		Code:              "0009",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		Receiver:          receiver,
		TemplateKey:       templatekey,
		TemplateParams:    templateParams,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewayIGGSMSTempl, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewayIGGSMSTempl, e, h)
}

//SendMsgIggTTSTemplate IGG语音模版消息 （消息代码：0010）比较特殊，不能设置自定义标题和正文，只支持固定模版发送
//
// baseParams 消息基本属性
//
// receiver 收件人
//
// templatekey 语音的模版KEY
//
// templateParams
func (a *MessageGateway) SendMsgIggTTSTemplate(baseParams MessageBaseParams, receiver IGGReceiver, templatekey string, templateParams interface{}) IGGError {
	query := &msgIggTTSTemplate{
		Code:              "0010",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		Receiver:          receiver,
		TtsCode:           templatekey,
		TtsParams:         templateParams,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewayIGGTTSTempl, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewayIGGTTSTempl, e, h)
}

//SendMsgIGGSupport IGG客服消息（消息代码：0011），发送给客服系统，参考运营后台：平台 > 报警系统 > 类型配置管理(4607)，不支持自定义标题
//
// baseParams 消息基本属性
//
// key 平台 > 报警系统 > 类型配置管理(4607) 下的KEY
//
// gameID 可选。游戏ID
//
// serverID 可选。服务器ID，-1表示全服
//
// level 设置报警等级
//
// contacts 可选。设置这个字段，会在客服消息正文后面附带联系人的联系方式，见10.4
//
// ignore 是否忽略报警
func (a *MessageGateway) SendMsgIGGSupport(baseParams MessageBaseParams, key string, gameID uint64, serverID string, level int16, contacts []IGGReceiver, ignore bool) IGGError {
	query := &msgIggSupport{
		Code:              "0011",
		AppID:             a.apiToken,
		MessageBaseParams: baseParams,
		Key:               key,
		GameID:            gameID,
		ServerID:          serverID,
		Level:             level,
		Contacts:          contacts,
		Ignore:            ignore,
	}
	s, e, h := a.gateway.Post(fmt.Sprintf("%v?app_id=%v", iggreq.MessageGatewaySendMsgAPIPath, a.apiToken), query)
	if !e.IsSuccess() {
		return msggwInternalErrorToSDKExceptton(gatewayIGGSupport, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return msggwInternalErrorToSDKExceptton(gatewayIGGSupport, e, h)
}

// 个人消息
type msgSingle struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	URL      string      `json:"url"`
	Receiver IGGReceiver `json:"receiver"`
}

// 群发个人消息
type msgMultiple struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	URL       string        `json:"url"`
	Receivers []IGGReceiver `json:"receivers"`
	Source    int           `json:"source"`
}

// 群组消息
type msgRoom struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	Room uint64 `json:"room"`
	At   string `json:"at"`
	Type int    `json:"type"`
}

// MyIGG消息
type msgMyIGG struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	URL      string      `json:"url"`
	Receiver IGGReceiver `json:"receiver"`
	Source   string      `json:"source"`
	Type     int         `json:"type"`
}

// IGG邮件消息
type msgMail struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	Receivers []IGGReceiver `json:"receivers"`
	Ccs       []IGGReceiver `json:"ccs"`
	Bccs      []IGGReceiver `json:"bccs"`
	HTML      bool          `json:"html"`
}

// 企业微信个人消息
type msgWXEntPerson struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	Receiver IGGReceiver `json:"receiver"`
	Source   string      `json:"source"`
}

// 企业微信部门消息
type msgWXEntDepartment struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	DepartmentID uint64 `json:"departmentID"`
	Source       string `json:"source"`
}

// 企业微信公司消息
type msgWXEntCompany struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	CompanyID uint64 `json:"companyID"`
	Source    string `json:"source"`
}

// IGG短信消息
type msgIggSMS struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	Receiver IGGReceiver `json:"receiver"`
}

// IGG短信模板消息
type msgIggSMSTemplate struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	Receiver       IGGReceiver `json:"receiver"`
	TemplateKey    string      `json:"templateKey"`
	TemplateParams interface{} `json:"templateParams"`
}

// IGG语音模版消息
type msgIggTTSTemplate struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	Receiver  IGGReceiver `json:"receiver"`
	TtsCode   string      `json:"ttsCode"`
	TtsParams interface{} `json:"ttsParams"`
}
type msgIggSupport struct {
	//Code 消息代码，不同的消息代码代表了不同的消息类型，不同消息类型可能需要设置额外的字段
	Code string `json:"code"`
	// AppID 应用ID，即API网关的TOKEN
	AppID string `json:"appID"`
	MessageBaseParams
	Key      string        `json:"key"`
	GameID   uint64        `json:"gameID"`
	ServerID string        `json:"serverID"`
	Level    int16         `json:"level"`
	Contacts []IGGReceiver `json:"contacts"`
	Ignore   bool          `json:"ignore"`
}

// GenQueryString generate query string.
func (q *msgSingle) GenQueryString() string {
	bts, _ := json.Marshal(q)
	//fmt.Println(string(bts))
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgMultiple) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgRoom) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgMyIGG) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgMail) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgWXEntPerson) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgWXEntDepartment) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgWXEntCompany) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgIggSMS) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgIggSMSTemplate) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgIggTTSTemplate) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// GenQueryString generate query string.
func (q *msgIggSupport) GenQueryString() string {
	bts, _ := json.Marshal(q)
	return string(bts)
}

// ContentType content-type of query data for POST request
func (q *MessageBaseParams) ContentType() string {
	return "application/json"
}

func msggwInternalErrorToSDKExceptton(bus APIBusiness, e *iggreq.InternalError, hdr http.Header) IGGError {
	code := IGGError{}
	if !e.IsSuccess() {
		code = NewIGGError(MSGGATEWAY, bus, int(e.Type))
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
