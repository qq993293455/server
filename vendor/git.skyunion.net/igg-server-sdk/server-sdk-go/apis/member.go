package apis

import (
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.skyunion.net/igg-server-sdk/server-sdk-go/iggreq"
)

//ExpirationInquery 过期查询参数
type ExpirationInquery struct {
	//player's iggid
	Iggid string
	//Hash code of device ID
	HashedUDID string
}

//ExpirationInqueryDelegate 过期查询委托
type ExpirationInqueryDelegate func(inquery ExpirationInquery) int64

// AuthRequest AcessToken 验证请求
type AuthRequest struct {
	Gameid      string
	Iggid       string
	Udid        string
	AccessToken string
}

// AccessToken 验证后得到的 access token 结构
type AccessToken struct {
	Gameid     string `json:"gid,omitempty"`
	Iggid      string `json:"sub,omitempty"`
	LoginType  string `json:"typ,omitempty"`
	HashedUDID string `json:"udid,omitempty"`
	Ttl        int64  `json:"ttl,omitempty"`
	Iat        int64  `json:"iat,omitempty"`
}

func (v *AccessToken) Expired() bool {
	return time.Now().Unix() > v.Iat+v.Ttl
}

type Account struct {
	gateway        *iggreq.GatewayRequest
	compatibleMode bool
}

const (
	memberVerify APIBusiness = iota + 2
	memberQueryMinVer
)

//NewAccountServiceDefault 创建新的会员接口(默认超时10秒)
func NewAccountServiceDefault() *Account {
	return NewAccountService(
		iggreq.ConfigInstance().GatewayConfig(),
		10)
}

//NewAccountService 创建新的会员接口
//
// gatewayConf API 网关配置消费者
//
// timeoutSecond API 网关接口调用超时
func NewAccountService(gatewayConf iggreq.GatewayConfiguration, timeoutSecond int) *Account {
	return &Account{
		gateway:        iggreq.NewGatewayRequest(gatewayConf, timeoutSecond),
		compatibleMode: false,
	}
}

type VeriyErrorCode int

const (
	AKFormat VeriyErrorCode = iota + 1
	AKDecode
	AKDecrypt
	AKRsaVerify
	AKExpired
	AKApiParam
	AKDevice
	GWVerifyNotAllowed
)

func createCGIError(code VeriyErrorCode, info string) IGGError {
	errImpl := IGGError{
		mod:           CGI,
		bus:           memberVerify,
		cat:           int(code),
		extensionCode: info,
	}
	return errImpl
}

func Base64UrlDecode(data string) ([]byte, error) {
	mod := len(data) % 4
	if mod > 0 {
		data += string(string("====")[:4-mod])
	}
	return base64.URLEncoding.DecodeString(data)
}

func createEncryptedToken(accessToken string) (token encodedToken, ex IGGError) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		ex = createCGIError(AKFormat, "Invalid AccessToken format.")
		return
	}
	_, err := Base64UrlDecode(parts[0])
	if err != nil {
		ex = createCGIError(AKDecode, "Failed to decode header.")
		return
	}
	_, err = Base64UrlDecode(parts[1])
	if err != nil {
		ex = createCGIError(AKDecode, "Failed to decode body."+err.Error())
		return
	}
	_, err = Base64UrlDecode(parts[2])
	if err != nil {
		ex = createCGIError(AKDecode, "Failed to decode signature.")
		return
	}
	token.Header = parts[0]
	token.Body = parts[1]
	token.Signature = parts[2]
	return
}

func Md5ToHexString(data []byte) string {
	bts := md5.Sum(data)
	return hex.EncodeToString(bts[:])
}

type jwk struct {
	N string `json:"n"`
	E string `json:"e"`
}

func (a *Account) SetCompatibleMode(enable bool) {
	a.compatibleMode = enable
}

//getMinimunVersion 从网关获取最小版本号
func (a *Account) getMinimunVersion(gameID string, inquery ExpirationInquery) (ver int64, ex IGGError) {
	query := iggreq.NewQueryMaker("")
	query.AddParam("hashed_udid", inquery.HashedUDID).AddParam("iggid", inquery.Iggid).AddParam("game_id", gameID)

	s, e, h := a.gateway.Get(iggreq.TokenMinimumVerAPIPath, query)
	traceid := e.TraceID
	if !e.IsSuccess() {
		ex = minimumVersionErrorToSDKExceptton(memberQueryMinVer, e, h)
		return 0, ex
	}

	d, e := iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	if !e.IsSuccess() {
		ex = minimumVersionErrorToSDKExceptton(memberQueryMinVer, e, h)
		if e.Type == iggreq.SDKEApi && e.Code == 201 {
			ex.cat = int(GWVerifyNotAllowed)
		}
		return 0, ex
	}

	ex = minimumVersionErrorToSDKExceptton(memberQueryMinVer, e, nil)
	var ok bool
	switch o := d.(type) {
	case map[string]interface{}:
		var buf interface{}
		var bad error
		if buf, ok = o["minimum_version"]; ok {
			ver, bad = strconv.ParseInt(buf.(string), 10, 64)
			if bad != nil {
				ok = false
			}
		}
	default:
	}
	if !ok {
		ex.cat = COMMON_ERROR_SERVICE
		ex.underlyingCode = strconv.FormatInt(int64(iggreq.SDKEApi), 10)
		ex.extensionCode = "cannot parse minimum_version"
	}
	return
}

//Verify Access token 验证
func (a *Account) Verify(
	request AuthRequest,
	aesKeyMap map[string]string,
	inqueryDelegate ExpirationInqueryDelegate) (info AccessToken, ex IGGError) {
	et, ex := createEncryptedToken(request.AccessToken)
	if !ex.IsSuccess() {
		return
	}

	akHdr, err := et.DecodHeader()
	if err != nil {
		ex = createCGIError(AKDecode, "Failed to parse header of AccessToken.")
		return
	}

	if _, ok := aesKeyMap[akHdr.Akid]; !ok {
		ex = createCGIError(AKDecrypt, "No matching AES key for id: "+akHdr.Akid)
		return
	}

	info, err = et.DecodeBody(aesKeyMap[akHdr.Akid])
	if err != nil {
		ex = createCGIError(AKDecode, "Failed to parse body of AccessToken."+err.Error())
		return
	}

	if !(request.Udid == "" && a.compatibleMode) {
		if Md5ToHexString([]byte(request.Udid))[:8] != info.HashedUDID {
			ex = createCGIError(AKDevice, "UDID not match.")
			return
		}
	}

	if request.Iggid != info.Iggid {
		ex = createCGIError(AKDevice, "iggid not match.")
		return
	}

	rsaKey, ok := certsCacheInstance().rsakeyMap[akHdr.Rkid]
	if certsCacheInstance().Expired() || (!ok && certsCacheInstance().CanUpdate()) {
		ex = reqireCerts(a.gateway, request.Gameid)
		if ex.IsSuccess() == false {
			return
		}
		rsaKey, ok = certsCacheInstance().rsakeyMap[akHdr.Rkid]
	}
	if !ok {
		createCGIError(AKRsaVerify, "No matching RSA key for id: "+akHdr.Rkid)
	}

	hash := crypto.SHA256.New()
	hash.Write([]byte(et.Body))

	err = rsa.VerifyPKCS1v15(&rsaKey, crypto.SHA256, hash.Sum(nil), et.DecodSignature())
	if err != nil {
		ex = createCGIError(AKRsaVerify, "Can not verify token signature.")
		return
	}

	// 验证有效期
	if time.Now().Unix() >= info.Iat+info.Ttl {
		ex = createCGIError(AKExpired, "Token has expired.")
		return
	}

	query := ExpirationInquery{
		Iggid:      info.Iggid,
		HashedUDID: info.HashedUDID,
	}

	var ver int64
	if inqueryDelegate != nil {
		ver = inqueryDelegate(query)
	} else {
		ver, ex = a.getMinimunVersion(request.Gameid, query)
	}
	if ver > info.Iat {
		ex = createCGIError(AKExpired, "Token has expired.")
	}
	return
}

func minimumVersionErrorToSDKExceptton(bus APIBusiness, e *iggreq.InternalError, hdr http.Header) IGGError {
	code := IGGError{
		mod: CGI,
	}
	if !e.IsSuccess() {
		switch e.Type {
		case iggreq.SDKELink:
			code.bus = commError
			code.cat = COMMON_ERROR_NETWORK
			break
		case iggreq.SDKEHttp:
			code.bus = commError
			code.cat = COMMON_ERROR_SERVICE
			break
		case iggreq.SDKEOther:
			code.bus = commError
			code.cat = COMMON_ERROR_OTHER
			break
		default:
			code.bus = bus
			code.cat = int(e.Type)
			break
		}

		//code = NewIGGError(int(CGI), bus, int(e.Type))
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
