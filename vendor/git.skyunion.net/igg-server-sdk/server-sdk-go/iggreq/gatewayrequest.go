package iggreq

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// AuthMethod authenticate metheod type
type AuthMethod int

const (
	// AM_HMAC_SHA1 authentication is generated using HMAC_SHA1
	AM_HMAC_SHA1 AuthMethod = iota
	// AM_HMAC_SHA256 authentication is generated using HMAC_SHA256
	AM_HMAC_SHA256
)

// GatewayRequest wrapping of gateway request
type GatewayRequest struct {
	GatewayConfiguration
	method  AuthMethod
	headers map[string]string
	timeout time.Duration
}

var gRand *rand.Rand

func init() {
	gRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// NewGatewayRequest create a new gateway request object
func NewGatewayRequest(gatewayConf GatewayConfiguration, timeoutSec int) *GatewayRequest {
	return &GatewayRequest{
		GatewayConfiguration: gatewayConf,
		method:               AM_HMAC_SHA256,
		timeout:              time.Duration(timeoutSec) * time.Second,
	}
}

// SetsHeader set http header name with value
func (gr *GatewayRequest) SetsHeader(name, value string) {
	if gr.headers == nil {
		gr.headers = make(map[string]string)
	}
	gr.headers[name] = value
}

// MultipartPost post form data
func (gr *GatewayRequest) MultipartPost(path, contentType string, body io.Reader) (res string, err *InternalError, hdr http.Header) {
	req, e := http.NewRequest("POST", gr.GatewayAddress+path, body)
	if e != nil {
		return res, InternalErrorFromError(SDKEOther, e, nil), nil
	}
	// Set headers
	req.Header = gr.genHeader(true, path, "", contentType)

	// Do
	cli := http.DefaultClient
	cli.Timeout = gr.timeout
	rsp, e := cli.Do(req)
	if e != nil {
		return res, InternalErrorFromError(SDKELink, e, req.Header), nil
	}
	defer rsp.Body.Close()
	hdr = rsp.Header
	buf, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return res, InternalErrorFromError(SDKELink, e, req.Header), hdr
	}
	if rsp.StatusCode != 200 {
		var content string
		for k, v := range rsp.Header {
			content += fmt.Sprintf("%s%v\r\n", k, v)
		}
		content += string(buf)
		return res, NewHttpInternalError(rsp.StatusCode, content, req.Header), hdr
	}

	traceid := ""
	if ids, ok := req.Header["X-Igg-Traceid"]; ok {
		traceid = ids[0]
	}
	return string(buf), NoError(traceid), hdr
}

// Post Do gateway requets using POST http method
func (gr *GatewayRequest) Post(path string, query QueryGeneratorInterface) (res string, ex *InternalError, hdr http.Header) {

	// New request
	body := query.GenQueryString()
	req, err := http.NewRequest("POST", gr.GatewayAddress+path, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return res, InternalErrorFromError(SDKEOther, err, nil), nil
	}

	// Set headers
	req.Header = gr.genHeader(true, path, body, query.ContentType())

	// Do
	cli := http.DefaultClient
	cli.Timeout = gr.timeout
	rsp, err := cli.Do(req)
	if err != nil {
		return res, InternalErrorFromError(SDKELink, err, req.Header), nil
	}
	defer rsp.Body.Close()
	hdr = rsp.Header
	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return res, InternalErrorFromError(SDKELink, err, req.Header), hdr
	}
	if rsp.StatusCode != 200 {
		var content string
		for k, v := range rsp.Header {
			content += fmt.Sprintf("%s%v\r\n", k, v)
		}
		content += string(buf)
		return res, NewHttpInternalError(rsp.StatusCode, content, req.Header), hdr
	}

	traceid := ""
	if ids, ok := req.Header["X-Igg-Traceid"]; ok {
		traceid = ids[0]
	}
	return string(buf), NoError(traceid), hdr
}

// Get Do gateway requets using GET http method
func (gr *GatewayRequest) Get(path string, query QueryGeneratorInterface) (res string, ex *InternalError, hdr http.Header) {
	buf, ex, hdr := gr.GetByte(path, query)
	return string(buf), ex, hdr
}

// GetByte Do gateway request using GET http method and return as byte array
func (gr *GatewayRequest) GetByte(path string, query QueryGeneratorInterface) (res []byte, ex *InternalError, hdr http.Header) {

	body := query.GenQueryString()
	resource := path + "?" + body
	// New request
	req, err := http.NewRequest("GET", gr.GatewayAddress+resource, nil)
	if err != nil {
		return res, InternalErrorFromError(SDKEOther, err, nil), nil
	}

	// Set headers
	req.Header = gr.genHeader(false, resource, "", "")

	// Do
	cli := http.DefaultClient
	cli.Timeout = gr.timeout
	rsp, err := cli.Do(req)
	if err != nil {
		return res, InternalErrorFromError(SDKELink, err, req.Header), nil
	}
	defer rsp.Body.Close()

	hdr = rsp.Header
	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return res, InternalErrorFromError(SDKELink, err, req.Header), hdr
	}
	if rsp.StatusCode != 200 {
		var content string
		for k, v := range rsp.Header {
			content += fmt.Sprintf("%s%v\r\n", k, v)
		}
		content += string(buf)
		return res, NewHttpInternalError(rsp.StatusCode, content, req.Header), hdr
	}
	traceid := ""
	if ids, ok := req.Header["X-Igg-Traceid"]; ok {
		traceid = ids[0]
	}

	return buf, NoError(traceid), hdr
}

func genTraceID() string {
	b := make([]byte, 16)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Read(b)
	return fmt.Sprintf("%x%x%x%x", b[:4], b[4:8], b[8:12], b[12:])
}

func genNonce() string {
	b := make([]byte, 8)
	gRand.Read(b)
	var data uint32
	binary.Read(bytes.NewBuffer(b), binary.BigEndian, &data)
	return fmt.Sprint(data)
}

func (gr *GatewayRequest) genHeader(post bool, path, body, contenttype string) (res http.Header) {
	res = http.Header{}
	for k, v := range gr.headers {
		res.Set(k, v)
	}

	var httpmethod string
	if post {
		httpmethod = "POST"
	} else {
		httpmethod = "GET"
	}

	// Reqired by gateway authentication
	res.Set("Date", formatTimeGMT())
	if body != "" {
		res.Set("Digest", "SHA-256="+base64Sha256Digest([]byte(body)))
	}
	res.Set("Authorization", gr.calcHeaderAuth(httpmethod, path, res))

	idx := strings.Index(gr.GatewayAddress, "://")
	if idx == -1 {
		idx = 0
	} else {
		idx += 3
	}
	if strings.Index(gr.GatewayAddress, "apis-dsa.igg.com") == idx || strings.Index(gr.GatewayAddress, "apis.igg.com") == idx {
		res.Set("X-IGG-NONCE", genNonce())
	}

	// Default HTTP headers
	res.Set("X-IGG-TRACEID", genTraceID())
	if _, ok := res["Content-Type"]; !ok && post {
		res.Set("Content-Type", contenttype)
	}
	if _, ok := res["User-Agent"]; !ok {
		res.Set("User-Agent", gConfig.GetUserAgent())
	}
	if _, ok := res["Accept"]; !ok {
		res.Set("Accept", "text/html;text/plain;text/json")
	}
	if _, ok := res["Connection"]; !ok {
		res.Set("Connection", "Close")
	}
	return res
}

func formatTimeGMT() string {
	loc := time.FixedZone("GMT", 0)
	return time.Now().In(loc).Format("Mon, 02 Jan 2006 15:04:05 GMT")
}

func base64Sha256Digest(data []byte) string {
	bodySHA := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(bodySHA[:])
}

func (gr *GatewayRequest) calcHeaderAuth(method, path string, headers http.Header) string {
	authstr := strings.Builder{}
	fields := strings.Builder{}
	// Add up all custom header
	for k, v := range headers {
		tmp := strings.ToLower(k)
		authstr.WriteString(fmt.Sprintf("%s: %s\n", tmp, v[0]))
		fields.WriteString(fmt.Sprintf("%s ", tmp))
	}

	// Add request line
	// 2020.5.28 王亮：网关鉴权这，原来进签名的string包括 http/1.1 这样的版本，现在去掉
	//authstr.WriteString(fmt.Sprintf("%s %s HTTP/1.1", method, path))
	authstr.WriteString(fmt.Sprintf("%s %s", method, path))
	fields.WriteString("request-line")

	// Hash
	var h hash.Hash
	var algorithm string
	switch gr.method {
	case AM_HMAC_SHA1:
		algorithm = "hmac-sha1"
		h = hmac.New(sha1.New, []byte(gr.GatewaySecret))
	case AM_HMAC_SHA256:
		algorithm = "hmac-sha256"
		h = hmac.New(sha256.New, []byte(gr.GatewaySecret))
	}

	h.Write([]byte(authstr.String()))
	hdrSign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	//
	res := strings.Builder{}
	res.WriteString(fmt.Sprintf(`hmac username="%s",`, gr.GatewayToken))
	res.WriteString(fmt.Sprintf(` algorithm="%s",`, algorithm))
	res.WriteString(fmt.Sprintf(` headers="%s",`, fields.String()))
	res.WriteString(fmt.Sprintf(` signature="%s"`, hdrSign))
	return res.String()
}

//ParseResultFromAPI test
func ParseResultFromAPI(rsp string) (interface{}, *InternalError) {
	if strings.Index(rsp, "\"errmsg\"") != -1 {
		return parseKVStoreJSON(rsp)
	} else if strings.Index(rsp, "\"errStr\"") != -1 {
		strBody := rsp[0 : strings.LastIndex(rsp, "}")+1]
		return parseOldJSON(strBody)
	} else if strings.Index(rsp, "\"message\"") != -1 {
		return parseJSON(rsp)
	} else {
		return parseNoneJSON(rsp)
	}
}

func getValueFromMap(obj interface{}, key string) interface{} {
	dict := obj.(map[string]interface{})
	if dict == nil {
		return nil
	}
	for k, v := range dict {
		if k == key {
			return v
		}
	}
	return nil
}
func parseJSON(body string) (data interface{}, ex *InternalError) {
	ex = &InternalError{
		Type: SDKEApi,
		Code: -1,
		Msg:  body,
	}

	var obj interface{}
	err := json.Unmarshal([]byte(body), &obj)
	if err != nil {
		return
	}

	objErr := getValueFromMap(obj, "error")
	if objErr == nil {
		return
	}
	switch code := getValueFromMap(objErr, "code").(type) {
	case float64:
		ex.Code = int(code)
	default:
		return
	}
	if ex.Code == 0 {
		ex.Type = SDKESuccess
		data = getValueFromMap(obj, "data")
	} else {
		switch msg := getValueFromMap(objErr, "message").(type) {
		case string:
			ex.Msg = msg
		}
	}
	return
}

func parseOldJSON(body string) (data interface{}, ex *InternalError) {

	ex = &InternalError{
		Type: SDKEApi,
		Code: -1,
		Msg:  body,
	}

	var obj interface{}
	err := json.Unmarshal([]byte(body), &obj)
	if err != nil {
		return
	}

	switch code := getValueFromMap(obj, "errCode").(type) {
	case float64:
		ex.Code = int(code)
	default:
		return
	}
	if ex.Code == 0 {
		ex.Type = SDKESuccess
		data = getValueFromMap(obj, "result")
	} else {
		switch msg := getValueFromMap(obj, "errStr").(type) {
		case string:
			ex.Msg = msg
		}
	}
	return
}

func parseKVStoreJSON(body string) (data interface{}, ex *InternalError) {
	ex = &InternalError{
		Type: SDKEApi,
		Code: -1,
		Msg:  body,
	}

	var obj interface{}
	err := json.Unmarshal([]byte(body), &obj)
	if err != nil {
		return
	}

	switch code := getValueFromMap(obj, "errcode").(type) {
	case float64:
		ex.Code = int(code)
	default:
		return
	}
	if ex.Code == 0 {
		ex.Type = SDKESuccess
		data = getValueFromMap(obj, "data")
	} else {
		switch msg := getValueFromMap(obj, "errmsg").(type) {
		case string:
			ex.Msg = msg
		}
	}
	return
}

func parseNoneJSON(body string) (data interface{}, ex *InternalError) {
	return nil, &InternalError{
		Type: SDKEApi,
		Code: -1,
		Msg:  body,
	}
}
