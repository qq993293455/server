package apis

import (
	"fmt"
)

/**
* SDK 内模块
 */
type APISDK_MODULE int

const (
	// 会员
	CGI APISDK_MODULE = iota + 11
	// 推送
	PUSH
	// 报警
	ALARM
	// KV 存储（战报）
	KVSTORE
	// IP 查询
	IPQUERY
	//消息网关
	MSGGATEWAY
)

type APIBusiness int

const (
	commError APIBusiness = 0
)

const (
	COMMON_ERROR_NETWORK   = 1
	COMMON_ERROR_SERVICE   = 2
	COMMON_ERROR_SUBSYSTEM = 4
	COMMON_ERROR_OTHER     = 9
)

//IGGError 错误
type IGGError struct {
	// 底层错误码
	underlyingCode string
	// 扩展错误码
	extensionCode string
	// 模块、业务、错误类型
	mod APISDK_MODULE
	bus APIBusiness
	cat int
	// traceIDs
	traceIDs []string
}

// NewIGGError 初始化错误码
func NewIGGError(mod APISDK_MODULE, bus APIBusiness, cat int) IGGError {
	return IGGError{
		mod: mod,
		bus: bus,
		cat: cat,
	}
}

func (e *IGGError) Code() int {
	if e.cat == 0 {
		return 0
	}
	return int(e.mod)*10000 + int(e.bus)*100 + int(e.cat)
}

func (e *IGGError) ToInteger() int {
	return e.Code()
}

func (e *IGGError) ToString() string {
	return fmt.Sprintf("%d", e.Code())
}

func (e *IGGError) ReadableCode() string {
	ret := e.ToString()
	if len(e.underlyingCode) == 0 {
		ret += "-0"
	} else {
		ret += "-" + e.underlyingCode
	}
	if len(e.extensionCode) > 0 {
		ret += "-" + e.extensionCode
	}
	return ret
}

func (e *IGGError) TraceIDs() []string {
	return e.traceIDs
}

func (e *IGGError) Dump() string {
	ret := e.ReadableCode()
	for _, trace := range e.traceIDs {
		ret += ", " + trace
	}
	return ret
}

func (e *IGGError) IsSuccess() bool {
	return e.cat == 0
}
