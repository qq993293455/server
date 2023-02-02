package errmsg

import (
	"fmt"
	"unsafe"

	"coin-server/common/proto/models"
	"coin-server/common/utils"
)

type ErrMsg models.Resp

func (this_ *ErrMsg) Error() string {
	if this_ == nil {
		return "nil"
	}
	s := ""
	if len(this_.StackStace) > 0 {
		s = fmt.Sprintf(`{"err_code":"%s","err_msg":"%s","err_internal_info":"%s","stack_stace":"%+v"}`,
			this_.ErrCode.String(), this_.ErrMsg, this_.ErrInternalInfo, (*utils.Stack)(unsafe.Pointer(&this_.StackStace)).StackTrace())
	} else {
		s = fmt.Sprintf(`{"err_code":"%s","err_msg":"%s","err_internal_info":"%s"}`,
			this_.ErrCode.String(), this_.ErrMsg, this_.ErrInternalInfo)
	}
	return s
}

func (this_ *ErrMsg) String() string {
	return this_.Error()
}

func (this_ *ErrMsg) WithStack() {
	stack := utils.Callers()
	this_.StackStace = *(*[]uint64)(unsafe.Pointer(stack))
}

func (this_ *ErrMsg) IsErrorNormal() bool {
	if this_ == nil {
		return false
	}
	return this_.ErrCode == models.ErrorType_ErrorNormal
}

func (this_ *ErrMsg) IsErrorProtocol() bool {
	if this_ == nil {
		return false
	}
	return this_.ErrCode == models.ErrorType_ErrorProtocol
}

func (this_ *ErrMsg) IsErrorPanic() bool {
	if this_ == nil {
		return false
	}
	return this_.ErrCode == models.ErrorType_ErrorPanic
}

func (this_ *ErrMsg) IsErrorDB() bool {
	if this_ == nil {
		return false
	}
	return this_.ErrCode == models.ErrorType_ErrorDB
}

func (this_ *ErrMsg) IsErrorNoLogin() bool {
	if this_ == nil {
		return false
	}
	return this_.ErrCode == models.ErrorType_ErrorNoLogin
}

func (this_ *ErrMsg) IsErrorNatsNoResponders() bool {
	if this_ == nil {
		return false
	}
	return this_.ErrCode == models.ErrorType_ErrorNatsNoResponders
}

const (
	InternalErrMsg        = "server_internal_error"
	AuthErrMsg            = "auth_failed_error"
	NotFoundHandlerErrMsg = "not_found_handler_error"
)

func NewErrorNatsNoResponders(msg string) *ErrMsg {
	e := &ErrMsg{
		ErrCode:         models.ErrorType_ErrorNatsNoResponders,
		ErrMsg:          "error_nats_no_responders",
		ErrInternalInfo: msg,
	}
	if openStack {
		caller := utils.Callers()
		e.StackStace = *(*[]uint64)(unsafe.Pointer(caller))
	}
	return e
}

func NewInternalErr(errInfo string) *ErrMsg {
	return NewNormalErr(InternalErrMsg, errInfo)
}

func NewNormalErr(errMsg string, errInfo string) *ErrMsg {
	e := &ErrMsg{
		ErrCode:         models.ErrorType_ErrorNormal,
		ErrMsg:          errMsg,
		ErrInternalInfo: errInfo,
	}
	if openStack {
		caller := utils.Callers()
		e.StackStace = *(*[]uint64)(unsafe.Pointer(caller))
	}
	return e
}

func NewProtocolErrorInfo(errInfo string) *ErrMsg {
	e := &ErrMsg{
		ErrCode:         models.ErrorType_ErrorProtocol,
		ErrMsg:          InternalErrMsg,
		ErrInternalInfo: errInfo,
	}
	if openStack {
		caller := utils.Callers()
		e.StackStace = *(*[]uint64)(unsafe.Pointer(caller))
	}
	return e
}

func NewProtocolError(err error) *ErrMsg {
	if err == nil {
		return nil
	}
	e := &ErrMsg{
		ErrCode:         models.ErrorType_ErrorProtocol,
		ErrMsg:          InternalErrMsg,
		ErrInternalInfo: err.Error(),
	}
	if openStack {
		caller := utils.Callers()
		e.StackStace = *(*[]uint64)(unsafe.Pointer(caller))
	}
	return e
}

func NewErrorPanic(errInfo interface{}) *ErrMsg {
	caller := utils.Callers()
	return &ErrMsg{
		ErrCode:         models.ErrorType_ErrorPanic,
		ErrMsg:          InternalErrMsg,
		ErrInternalInfo: fmt.Sprintf("%v", errInfo),
		StackStace:      *(*[]uint64)(unsafe.Pointer(caller)),
	}
}

func NewErrorPanicWith(e *ErrMsg, errInfo interface{}) *ErrMsg {
	if e == nil {
		return nil
	}
	caller := utils.Callers()
	e.ErrCode = models.ErrorType_ErrorPanic
	e.ErrMsg = InternalErrMsg
	e.ErrInternalInfo = fmt.Sprintf("%v", errInfo)
	e.StackStace = *(*[]uint64)(unsafe.Pointer(caller))
	return e
}

func NewErrorDB(err error) *ErrMsg {
	if err == nil {
		return nil
	}
	e, ok := err.(*ErrMsg)
	if ok {
		return e
	}
	return &ErrMsg{
		ErrCode:         models.ErrorType_ErrorDB,
		ErrMsg:          InternalErrMsg,
		ErrInternalInfo: err.Error(),
	}
}

func NewErrorDBInfo(errInfo string) *ErrMsg {
	e := &ErrMsg{
		ErrCode:         models.ErrorType_ErrorDB,
		ErrMsg:          InternalErrMsg,
		ErrInternalInfo: errInfo,
	}
	if openStack {
		caller := utils.Callers()
		e.StackStace = *(*[]uint64)(unsafe.Pointer(caller))
	}
	return e
}

func NewErrorNoLogin(err error) *ErrMsg {
	if err == nil {
		return nil
	}
	e, ok := err.(*ErrMsg)
	if ok {
		return e
	}
	e = &ErrMsg{
		ErrCode:         models.ErrorType_ErrorNoLogin,
		ErrMsg:          AuthErrMsg,
		ErrInternalInfo: err.Error(),
	}

	if openStack {
		caller := utils.Callers()
		e.StackStace = *(*[]uint64)(unsafe.Pointer(caller))
	}
	return e
}

func MustError(err error) {
	if err != nil {
		em, ok := err.(*ErrMsg)
		if ok {
			if em != nil {
				panic(em)
			}
		} else {
			panic(err)
		}
	}
}

func Must(e *ErrMsg) {
	if e != nil {
		panic(e)
	}
}
