package utils

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"coin-server/common/values"
	"coin-server/pikaviewer/global"

	"github.com/gin-gonic/gin"
)

const (
	DebugKey              = "DEBUG"
	DebugSignPlaintextKey = "SIGN"
	DebugSignKey          = "PLAINTEXT"
)

type Response struct {
	Ctx *gin.Context
	CodeInfo
	Data interface{}
}

type CodeInfo struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (c *CodeInfo) Error() string {
	return c.Msg
}

func NewResponse(ctx *gin.Context) *Response {
	return &Response{
		Ctx: ctx,
		CodeInfo: CodeInfo{
			Code: 0,
			Msg:  "success",
		},
		Data: nil,
	}
}

func (r *Response) Send(data ...interface{}) {
	var _data interface{}
	if len(data) > 0 {
		_data = data[0]
	}
	if _data == nil {
		r.Ctx.JSON(http.StatusOK, gin.H{
			"code": r.Code,
			"msg":  r.Msg,
		})
		return
	}
	typ := reflect.TypeOf(_data).String()
	var obj gin.H
	if typ == "*utils.CodeInfo" {
		_code := _data.(*CodeInfo)
		obj = gin.H{
			"code": _code.Code,
			"msg":  _code.Msg,
		}
	} else if strings.Index(typ, "Error") > -1 || strings.Index(typ, "error") > -1 {
		err, ok := _data.(error)
		msg := "server internal error"
		obj = gin.H{
			"code": 1,
		}
		if ok {
			msg = err.Error()
		} else {
			obj["details"] = _data
		}
		obj["msg"] = msg

	} else {
		obj = gin.H{
			"code": r.Code,
			"msg":  r.Msg,
			"data": _data,
		}
	}

	if _, ok := r.Ctx.Get(DebugKey); ok && global.Config.EnableDebug {
		plaintext, ok1 := r.Ctx.Get(DebugSignPlaintextKey)
		sign, ok2 := r.Ctx.Get(DebugSignKey)
		if !ok1 {
			plaintext = "debug plaintext not found"
		}
		if !ok2 {
			sign = "debug sign not found"
		}
		obj["debug"] = gin.H{
			"plaintext": plaintext,
			"sign":      sign,
		}
	}
	r.Ctx.JSON(http.StatusOK, obj)
}

func NewDefaultError(err ...error) *CodeInfo {
	if len(err) > 0 {
		return &CodeInfo{Code: 1, Msg: err[0].Error()}
	} else {
		return &CodeInfo{Code: 1, Msg: "failure"}
	}
}

func NewDefaultErrorWithMsg(msg ...string) *CodeInfo {
	if len(msg) > 0 {
		return &CodeInfo{Code: 1, Msg: msg[0]}
	} else {
		return &CodeInfo{Code: 1, Msg: "failure"}
	}
}

func NewErrWithCodeMsg(code *CodeInfo, msg string) *CodeInfo {
	code.Msg = msg
	return code
}

func NewErrWithMsg(msg ...string) error {
	_msg := "failure"
	if len(msg) > 0 {
		_msg = msg[0]
	}
	return errors.New(_msg)
}

func GetRoleId(key string) (values.RoleId, error) {
	listByColon := strings.Split(key, ":")
	listByDot := strings.Split(key, ".")
	var roleId values.RoleId
	if len(listByColon) >= 2 {
		roleId = listByColon[1]
	}
	if roleId == "" && len(listByDot) >= 2 {
		roleId = listByDot[1]
	}
	if roleId == "" {
		return "", errors.New("未解析到role id")
	}
	return roleId, nil
}
