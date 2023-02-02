package core

import (
	"fmt"

	"coin-server/common/values"
)

const DefaultErrorCode values.Integer = 1

var (
	ErrInvalidRequestLength = NewErrorMsg(DefaultErrorCode, "invalid request length")
)

func MustError(err error) {
	if err != nil {
		panic(err)
	}
}

type ErrorMsg struct {
	Code values.Integer
	Msg  string
}

func NewErrorMsg(code values.Integer, msg string) *ErrorMsg {
	return &ErrorMsg{Code: code, Msg: msg}
}

func NewDefaultCodeMsg(msg string) *ErrorMsg {
	return &ErrorMsg{Code: DefaultErrorCode, Msg: msg}
}

// 实现error interface
func (e *ErrorMsg) Error() string {
	return fmt.Sprintf("code:%d msg:%s", e.Code, e.Msg)
}

func (e *ErrorMsg) GetErrorCode() values.Integer {
	return e.Code
}

func (e *ErrorMsg) GetErrorMsg() string {
	return e.Msg
}
