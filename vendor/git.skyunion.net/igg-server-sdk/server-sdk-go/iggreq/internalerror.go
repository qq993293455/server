package iggreq

import (
	"net/http"
)

// ExceptionType SDK error type
type ExceptionType int

const (
	// SDKESuccess success get api result
	SDKESuccess ExceptionType = 0
	//SDKELink Network or system error
	SDKELink = 96
	// SDKEHttp HTTP server error
	SDKEHttp = 97
	// SDKEOther Other exceptions
	SDKEOther = 98
	// SDKEApi API return error
	SDKEApi = 99
)

// InternalError SDK callback result
type InternalError struct {
	Type    ExceptionType
	Code    int
	Msg     string
	TraceID string
}

func NoError(traceid string) *InternalError {
	return &InternalError{
		Type:    SDKESuccess,
		TraceID: traceid,
	}
}

func (e InternalError) IsSuccess() bool {
	return e.Type == SDKESuccess
}

// InternalErrorFromError Create SDKException from error
func InternalErrorFromError(ty ExceptionType, err error, hdr http.Header) *InternalError {
	ex := &InternalError{
		Type: ty,
	}
	traceID := ""
	if hdr != nil {
		if values, ok := hdr["X-IGG-TRACEID"]; ok {
			traceID = values[0]
		}
	}
	if err != nil {
		ex.Code = -1
		ex.Msg = err.Error()
		ex.TraceID = traceID
	}
	return ex
}

// InternalErrorFromResult Create SDKException from error
func InternalErrorFromResult(ty ExceptionType, code int, err string) *InternalError {
	return &InternalError{
		Type: ty,
		Code: code,
		Msg:  err,
	}
}

// NewHttpInternalError Create SDKException from error
func NewHttpInternalError(code int, content string, hdr http.Header) *InternalError {
	traceID := ""
	if hdr != nil {
		if values, ok := hdr["X-IGG-TRACEID"]; ok {
			traceID = values[0]
		}
	}
	return &InternalError{
		Type:    SDKEHttp,
		Code:    code,
		Msg:     content,
		TraceID: traceID,
	}
}
