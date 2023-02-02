package apis

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"git.skyunion.net/igg-server-sdk/server-sdk-go/iggreq"
)

//KVStorage KV存储接口
type KVStorage struct {
	gateway *iggreq.GatewayRequest
}

type kvQueryMaker struct {
	iggreq.QueryMaker
}

const (
	kvbPush APIBusiness = iota
	kvbCDNPush
	kvbCDNDel
	kvbKVPush
	kvbKVPull
)

func (q *kvQueryMaker) GenQueryString() string {
	// 排序
	sort.Slice(q.Params, func(i, j int) bool {
		return strings.Compare(q.Params[i].Name, q.Params[j].Name) < 0
	})

	gameid := uint64(0)
	file := ""
	// 构造
	res := strings.Builder{}
	for _, p := range q.Params {
		res.WriteString(fmt.Sprintf("%s=%s&", p.Name, url.QueryEscape(p.Value)))
		if p.Name == "g_id" {
			gameid, _ = strconv.ParseUint(p.Value, 10, 64)
		} else if p.Name == "file" {
			file = p.Value
		}
	}

	// 签名
	if q.SignKey != "" {
		res.WriteString(fmt.Sprintf("sign=%v", sign(gameid, file)))
	}

	return strings.TrimSuffix(res.String(), "&")
}

// NewKVStorageServiceDefault 新建一个KV存储服务对象，使用默认的token及secret （默认超时3分钟）
func NewKVStorageServiceDefault() *KVStorage {
	return NewKVStorageService(iggreq.ConfigInstance().GatewayConfig(), 60*3)
}

//NewKVStorageService 创建一个KV存储服务对象
//
// apiToken 消费者token
//
// apiSecret 消费者secret
func NewKVStorageService(gatewayConf iggreq.GatewayConfiguration, timeoutSecond int) *KVStorage {
	return &KVStorage{
		gateway: iggreq.NewGatewayRequest(gatewayConf, timeoutSecond),
	}
}

// Push 非永久存储推送至as3，定期清理
//
// gameid 游戏ID
//
// video 文件名, 只接受.gz扩展名
//
// day 保留期限3, 8, 30
//
// filestream 文件内容
//
// timestamp 指定时间戳（0 不指定）
//
// 返回 下载地址
func (kv *KVStorage) Push(gameid uint64, video string, day int, filestream io.Reader, timestamp int64) (string, IGGError) {
	s, e, h := kv.internalPush(iggreq.KVStorageAPIPushPath, 3, gameid, video, day, filestream, timestamp)
	if !e.IsSuccess() {
		return "", kvstoreInternalErrorToSDKExceptton(kvbPush, e, h)
	}
	traceid := e.TraceID
	d, e := iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	switch o := d.(type) {
	case string:
		return o, kvstoreInternalErrorToSDKExceptton(kvbPush, e, h)
	}
	return "", kvstoreInternalErrorToSDKExceptton(kvbPush, e, h)
}

// KVPush 永久存储，下载次数少且有更新需求，可使用此接口
//
// gameid 游戏ID
//
// video 文件名, 只接受.gz扩展名
//
// filestream 文件内容
//
// 返回 json 两个字段：errcode/errmsg
func (kv *KVStorage) KVPush(gameid uint64, video string, filestream io.Reader) IGGError {
	s, e, h := kv.internalPush(iggreq.KVStorageAPIKVPushPath, 2, gameid, video, -1, filestream, 0)
	if !e.IsSuccess() {
		return kvstoreInternalErrorToSDKExceptton(kvbKVPush, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid

	return kvstoreInternalErrorToSDKExceptton(kvbKVPush, e, h)
}

// CDNPush 永久存储推送至CDN,下载次数多的场景使用此接口
//
// gameid 游戏ID
//
// video 文件名, 只接受.gz扩展名
//
// filestream 文件内容
//
// timestamp 指定时间戳（0 不指定）
//
// 返回 json 三个字段：errcode/errmsg/data。下载地址在返回的 data 字段中，
func (kv *KVStorage) CDNPush(gameid uint64, video string, filestream io.Reader, timestamp int64) (string, IGGError) {
	s, e, h := kv.internalPush(iggreq.KVStorageAPICDNPushPath, 3, gameid, video, -1, filestream, timestamp)
	if !e.IsSuccess() {
		return "", kvstoreInternalErrorToSDKExceptton(kvbCDNPush, e, h)
	}
	traceid := e.TraceID
	d, e := iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	switch o := d.(type) {
	case string:
		return o, kvstoreInternalErrorToSDKExceptton(kvbCDNPush, e, h)
	}
	return "", kvstoreInternalErrorToSDKExceptton(kvbCDNPush, e, h)
}

// CDNDelete 永久存储从CDN中删除，下载次数多的场景使用此接口
func (kv *KVStorage) CDNDelete(gameid uint64, video string) IGGError {
	query := &kvQueryMaker{
		QueryMaker: iggreq.QueryMaker{
			SignKey: iggreq.KVStorageSecCode,
		},
	}
	query.AddParam("v", 3)
	query.AddParam("g_id", gameid)
	query.AddParam("file", video)
	s, e, h := kv.gateway.Post(iggreq.KVStorageAPICDNDeletePath, query)
	if !e.IsSuccess() {
		return kvstoreInternalErrorToSDKExceptton(kvbCDNDel, e, h)
	}
	traceid := e.TraceID
	_, e = iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	return kvstoreInternalErrorToSDKExceptton(kvbCDNDel, e, h)
}

// KVPull 永久存储，下载
//
// gameid 游戏ID
//
// video 文件名, 只接受.gz扩展名
func (kv *KVStorage) KVPull(gameid uint64, video string) ([]byte, IGGError) {
	query := &kvQueryMaker{
		QueryMaker: iggreq.QueryMaker{},
	}

	query.AddParam("g_id", gameid)
	query.AddParam("video", video)

	s, e, h := kv.gateway.GetByte(iggreq.KVStorageAPIKVPullPath, query)
	return s, kvstoreInternalErrorToSDKExceptton(kvbKVPull, e, h)
}

func (kv *KVStorage) internalPush(path string, ver int, gameid uint64, video string, day int, filestream io.Reader, timestamp int64) (string, *iggreq.InternalError, http.Header) {
	buf := &bytes.Buffer{}

	mp := multipart.NewWriter(buf)
	mp.WriteField("v", fmt.Sprint(ver))
	mp.WriteField("g_id", fmt.Sprint(gameid))
	if timestamp > 0 {
		mp.WriteField("timestamp", fmt.Sprint(timestamp))
	}
	if day > 0 {
		mp.WriteField("day", fmt.Sprint(day))
	}
	mp.WriteField("sign", sign(gameid, video))
	mp.CreateFormFile("video", video)

	contentType := mp.FormDataContentType()
	fields := make([]byte, buf.Len())
	buf.Read(fields)
	mp.Close()
	//part: latest boundary
	//when multipart closed, latest boundary is added
	lastBoundary := make([]byte, buf.Len())
	buf.Read(lastBoundary)

	//use pipe to pass request
	rd, wr := io.Pipe()
	defer rd.Close()

	go pipeWriter(wr, fields, lastBoundary, filestream)

	return kv.gateway.MultipartPost(path, contentType, rd)
}

func sign(gameid uint64, video string) (ret string) {
	qstr := fmt.Sprintf("%v%v%v", gameid, video, iggreq.KVStorageSecCode)
	sign := md5.Sum([]byte(qstr))
	ret = hex.EncodeToString(sign[:])
	//fmt.Println(qstr, ret)
	return
}

func pipeWriter(wr *io.PipeWriter, fields, lastBoundary []byte, stream io.Reader) {
	defer wr.Close()

	wr.Write(fields)

	//write file
	buf := make([]byte, 2048)
	for {
		n, e := stream.Read(buf)
		if e != nil {
			break
		}
		_, e = wr.Write(buf[:n])
		if e != nil {
			break
		}
	}
	//write boundary
	_, _ = wr.Write(lastBoundary)
}

func kvstoreInternalErrorToSDKExceptton(bus APIBusiness, e *iggreq.InternalError, hdr http.Header) IGGError {
	code := IGGError{}
	if !e.IsSuccess() {
		code = NewIGGError(KVSTORE, bus, int(e.Type))
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
