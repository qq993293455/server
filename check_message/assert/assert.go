package assert

import (
	"errors"
	"fmt"

	"coin-server/check_message/core"
	"coin-server/common/errmsg"
)

func Equal(ctx *core.RoleContext, expected, actual interface{}, msgAndArgs ...interface{}) {
	if !ObjectsAreEqual(expected, actual) {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert Equal Fail! userId:%s roleId:%s expected:%+v actual：%+v %s",
			ctx.UserId, ctx.RoleId, expected, actual, messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(str)
	}
}

func NotEqual(ctx *core.RoleContext, expected, actual interface{}, msgAndArgs ...interface{}) {
	if ObjectsAreEqual(expected, actual) {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert NotEqual Fail! userId:%s roleId:%s expected:%+v actual：%+v %s",
			ctx.UserId, ctx.RoleId, expected, actual, messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(str)
	}
}

func Empty(ctx *core.RoleContext, object interface{}, msgAndArgs ...interface{}) {
	if !isEmpty(object) {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert Empty Fail! userId:%s roleId:%s val:%+v %s", ctx.UserId, ctx.RoleId,
			object, messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(str)
	}
}

func NotEmpty(ctx *core.RoleContext, object interface{}, msgAndArgs ...interface{}) {
	if isEmpty(object) {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert NotEmpty Fail! userId:%s roleId:%s val:%+v %s", ctx.UserId, ctx.RoleId,
			object, messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(str)
	}
}

func Nil(ctx *core.RoleContext, val interface{}, msgAndArgs ...interface{}) {
	if !isNil(val) {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert Nil Fail! userId:%s roleId:%s val:%+v %s", ctx.UserId, ctx.RoleId,
			val, messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(str)
	}
}

func NotNil(ctx *core.RoleContext, val interface{}, msgAndArgs ...interface{}) {
	if isNil(val) {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert NotNil Fail! userId:%s roleId:%s val:%+v %s", ctx.UserId, ctx.RoleId,
			val, messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(str)
	}
}

func True(ctx *core.RoleContext, val bool, msgAndArgs ...interface{}) {
	if !val {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert True Fail! userId:%s roleId:%s val:%+v %s", ctx.UserId, ctx.RoleId,
			val, messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(str)
	}
}

func False(ctx *core.RoleContext, val bool, msgAndArgs ...interface{}) {
	if val {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert False Fail! userId:%s roleId:%s val:%+v %s", ctx.UserId, ctx.RoleId,
			val, messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(str)
	}
}

func Error(ctx *core.RoleContext, err error, msgAndArgs ...interface{}) {
	e, ok := err.(*errmsg.ErrMsg)
	if err == nil || (ok && e == nil) {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert Error Fail! userId:%s roleId:%s %s", ctx.UserId, ctx.RoleId,
			messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(err)
	}
}

func ErrorIs(ctx *core.RoleContext, err1, err2 error, msgAndArgs ...interface{}) {
	e1, ok1 := err1.(*errmsg.ErrMsg)
	e2, ok2 := err2.(*errmsg.ErrMsg)
	if (ok1 != ok2) || (ok1 == ok2 && !ok1 && !errors.Is(err1, err2)) || (ok1 == ok2 && ok1 && (e1.ErrCode != e2.ErrCode ||
		e1.ErrMsg != e2.ErrMsg)) {
		f, _ := getCallerFrame(3)
		str := fmt.Sprintf("Assert ErrorIs Fail! userId:%s roleId:%s  err1:%+v err2:%+v %s", ctx.UserId, ctx.RoleId,
			err1, err2, messageFromMsgAndArgs(msgAndArgs))
		ctx.RecordFailure(f.Function, 0, str)
		panic(str)
	}
}
