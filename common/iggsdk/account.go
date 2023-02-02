package iggsdk

import (
	"fmt"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"

	"git.skyunion.net/igg-server-sdk/server-sdk-go/apis"
)

type ApiAccount interface {
	Verify(c *ctx.Context, ar apis.AuthRequest, f func(c *ctx.Context, key string) int64) *errmsg.ErrMsg
}

type apiAccount struct {
	ac        *apis.Account
	aesKeyMap map[string]string
}

var accountIns *apiAccount

func InitAccountIns() {
	accountIns = &apiAccount{ac: apis.NewAccountServiceDefault(), aesKeyMap: map[string]string{"20220926": "rRXraBoOP4kArfjU"}}
}

func GetAccountIns() ApiAccount {
	return accountIns
}

func (aa *apiAccount) Verify(c *ctx.Context, ar apis.AuthRequest, f func(c *ctx.Context, key string) int64) *errmsg.ErrMsg {
	info, err := aa.ac.Verify(ar, aa.aesKeyMap, func(inquery apis.ExpirationInquery) int64 {
		return f(c, inquery.Iggid+"_"+inquery.HashedUDID)
	})
	fmt.Println(info, err)
	if !err.IsSuccess() {
		return errmsg.NewInternalErr(err.Dump())
	}
	return nil
}

func ConvertToIGGId(userId string) int64 {
	ret, err := strconv.Atoi(userId)
	if err != nil {
		return 0
	}
	return int64(ret)
}
