package apis

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"git.skyunion.net/igg-server-sdk/server-sdk-go/iggreq"
)

// PushType 推送的类型
type PushType int

const (
	// PT_Multicast 群发（技术部内部使用）
	PT_Multicast PushType = iota + 1
	//PT_Unicast 单发
	PT_Unicast
)

//PushSendMsgOptional 推送接口的可选参数
type PushSendMsgOptional struct {
	// 额外的键值数组,JSON字符串
	Data string
	// 消息标题
	Title string
	//指定发送用户的设备ID列表, 群发时与参数 `uidList` 二选一。一行一个
	Devid_list string
	//群推的批次号
	Mass_id string
	// 是否显示
	Display bool
	//IOS
	IOS_badge string
}

//PushSendMsgLoginOptional 推送登录失败的可选参数
type PushSendMsgLoginOptional struct {
	// 额外的键值数组,JSON字符串
	Data string
	// 消息标题
	Title string
	// 要推送设备的ADID(UDID)
	D_adid string
}

//Push 推送
type Push struct {
	gateway *iggreq.GatewayRequest
	signkey string
}

const (
	pushSend APIBusiness = iota
	pushSendLogin
)

// NewPushServiceDefault 创建推送接口实例(默认超时10秒)
func NewPushServiceDefault() *Push {
	return NewPushService(iggreq.ConfigInstance().GatewayConfig(), 10)
}

// 2020.6.3 王亮：我刚才问了郁菲，他报警和推送的原来的签名都不用了
//func NewPush(apiToken, apiSecret, signKey string) *Push {

// NewPushService 创建推送接口实例
func NewPushService(gatewayConf iggreq.GatewayConfiguration, timeoutSeconds int) *Push {
	return &Push{
		gateway: iggreq.NewGatewayRequest(gatewayConf, timeoutSeconds),
		//signkey: signKey,
	}
}

type pushQueryMaker struct {
	*iggreq.QueryMaker
}

// GenQueryString generate query string.
// The final query string need an additional field named "sign".
// It's calculated as folowing:
// First generate query string from parameters without URL encoding.
// Next adds the string with security code.
// Then Do MD5-HMAC sum to the result string, and get HEX string form of sum
// Finally append the field "sign" with that string
func (q *pushQueryMaker) GenQueryString() string {
	// 排序
	sort.Slice(q.Params, func(i, j int) bool {
		return strings.Compare(q.Params[i].Name, q.Params[j].Name) < 0
	})

	// 构造
	sb := strings.Builder{}
	res := strings.Builder{}
	for _, p := range q.Params {
		sb.WriteString(fmt.Sprintf("%s", p.Value))
		res.WriteString(fmt.Sprintf("%s=%s&", p.Name, url.QueryEscape(p.Value)))
	}

	// 签名 md5-hmac
	if q.SignKey != "" {
		query := strings.TrimSuffix(sb.String(), "&")
		h := hmac.New(md5.New, []byte(q.SignKey))
		h.Write([]byte(query))
		sum := h.Sum(nil)
		res.WriteString(fmt.Sprintf("_signature=%v&", hex.EncodeToString(sum[:])))
		res.WriteString(fmt.Sprintf("_timestamp=%v", time.Now().Unix()))
	}

	return strings.TrimSuffix(res.String(), "&")
}

//SendMsg 推送消息
//
// gid	游戏ID
//
// pt	群发与单发
//
// contant	推送内容
//
// iggid_list	推送的会员ID列表(一行一个)。 单发时必传，群发时与
//					设备列表二选一，@see PushSendMsgOptional.devid_list
//
// opts 可选参数
func (p *Push) SendMsg(gid int64, pt PushType, content, iggid_list string, opts *PushSendMsgOptional) IGGError {
	query := &pushQueryMaker{
		iggreq.NewQueryMaker(p.signkey),
	}
	query.AddParam("g_id", gid)
	query.AddParam("m_push_type", pt)
	query.AddParam("m_msg", content)

	//if pt == PT_Unicast && uidList == "" {
	//	return "", fmt.Errorf("Invalid arguments")
	//}
	//if pt == PT_Multicast && (uidList == "" && (opts == nil || opts.devid_list == "")) {
	//	return "", fmt.Errorf("Invalid arguments")
	//}

	if pt == PT_Unicast && iggid_list != "" {
		query.AddParam("m_iggid_file", iggid_list)
	}

	if opts != nil {
		if opts.Data != "" {
			query.AddParam("m_data", opts.Data)
		}
		if opts.Title != "" {
			query.AddParam("m_title", opts.Title)
		}
		if pt == PT_Multicast && iggid_list == "" {
			query.AddParam("m_regid_file", opts.Devid_list)
		}
		if opts.Mass_id != "" {
			query.AddParam("m_mass_id", opts.Mass_id)
		}
		if opts.Display {
			query.AddParam("m_display", 1)
		} else {
			query.AddParam("m_display", 0)
		}
		if opts.IOS_badge != "" {
			query.AddParam("m_ios_badge", opts.IOS_badge)
		}

	}
	s, e, h := p.gateway.Post(iggreq.PushSendMsgAPIPath, query)
	if !e.IsSuccess() {
		return pushInternalErrorToSDKExceptton(pushSend, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return pushInternalErrorToSDKExceptton(pushSend, e, h)
}

//SendMsgForLogin 推送异常登录消息
//
// gid	游戏ID
//
// uid	会员ID
//
// contant	推送内容
//
// opts 可选参数
func (p *Push) SendMsgForLogin(gid, uid int64, content string, opts *PushSendMsgLoginOptional) IGGError {
	query := &pushQueryMaker{
		iggreq.NewQueryMaker(p.signkey),
	}
	query.AddParam("g_id", gid)
	query.AddParam("iggid", gid)
	query.AddParam("m_msg", content)

	if opts != nil {
		if opts.Data != "" {
			query.AddParam("m_data", opts.Data)
		}
		if opts.Title != "" {
			query.AddParam("m_title", opts.Title)
		}
		if opts.D_adid != "" {
			query.AddParam("d_adid", opts.D_adid)
		}
	}
	s, e, h := p.gateway.Post(iggreq.PushSendMsgLoginAPIPath, query)
	if !e.IsSuccess() {
		return pushInternalErrorToSDKExceptton(pushSendLogin, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return pushInternalErrorToSDKExceptton(pushSendLogin, e, h)
}

func pushInternalErrorToSDKExceptton(bus APIBusiness, e *iggreq.InternalError, hdr http.Header) IGGError {
	code := IGGError{}
	if !e.IsSuccess() {
		code = NewIGGError(PUSH, bus, int(e.Type))
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
