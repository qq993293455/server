package iggsdk

import (
	"coin-server/common/consulkv"
	"coin-server/common/errmsg"
	"coin-server/common/utils"
	"coin-server/common/values"
	"git.skyunion.net/igg-server-sdk/server-sdk-go/apis"
)

type ApiPush interface {
	SendMsg(gameId int64, iggId string, content string) *errmsg.ErrMsg
	SendMsgForLogin(gameId, iggId int64, content string) *errmsg.ErrMsg
	GetLag(gameId values.Integer) string
}

type apiPush struct {
	ap  *apis.Push
	cfg map[values.Integer]string
}

var pushIns *apiPush

func InitPushIns(cnf *consulkv.Config) {
	cfg := &SdkGameId{}
	utils.Must(cnf.Unmarshal("sdk-game-id", cfg))
	cfgMap := map[values.Integer]string{}
	cfgMap[cfg.AndroidEN] = "en"
	cfgMap[cfg.IOSCN] = "cn"
	cfgMap[cfg.IOSEN] = "en"
	cfgMap[cfg.AndroidTw] = "hk"
	pushIns = &apiPush{
		ap:  apis.NewPushServiceDefault(),
		cfg: cfgMap,
	}
}

func GetPushIns() ApiPush {
	return pushIns
}

func (ap *apiPush) GetLag(gameId values.Integer) string {
	return ap.cfg[gameId]
}

func (ap *apiPush) SendMsg(gameId int64, iggId string, content string) *errmsg.ErrMsg {
	err := ap.ap.SendMsg(gameId, apis.PT_Unicast, content, iggId, nil)
	if !err.IsSuccess() {
		return errmsg.NewInternalErr(err.Dump())
	}
	return nil
}

func (ap *apiPush) SendMsgForLogin(gameId, iggId int64, content string) *errmsg.ErrMsg {
	err := ap.ap.SendMsgForLogin(gameId, iggId, content, nil)
	if !err.IsSuccess() {
		return errmsg.NewInternalErr(err.Dump())
	}
	return nil
}
