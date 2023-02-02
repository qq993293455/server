package iggsdk

import (
	"coin-server/common/consulkv"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/utils"
	"git.skyunion.net/igg-server-sdk/server-sdk-go/apis"
	"go.uber.org/zap"
)

type ApiAlarm interface {
	Send(content string) *errmsg.ErrMsg
	SendRes(content string) *errmsg.ErrMsg
}

type apiAlarm struct {
	aa  *apis.Alarm
	log *logger.Logger
	cfg *SdkTokenCfg
}

var alarmIns *apiAlarm

func InitAlarmIns(log *logger.Logger, cnf *consulkv.Config) {
	cfg := &SdkTokenCfg{}
	utils.Must(cnf.Unmarshal("sdk-token", cfg))
	alarmIns = &apiAlarm{
		aa:  apis.NewAlarmServiceDefault(),
		log: log,
		cfg: cfg,
	}
}

func GetAlarmIns() ApiAlarm {
	return alarmIns
}

func (aa *apiAlarm) Send(content string) *errmsg.ErrMsg {
	aa.log.Info("IGG SDK alarm", zap.String("token", aa.cfg.BattleToken))
	if aa.cfg.BattleToken == "" {
		return nil
	}
	err := aa.aa.Send(aa.cfg.BattleToken, content, nil)
	if !err.IsSuccess() {
		return errmsg.NewInternalErr(err.Dump())
	}
	return nil
}

func (aa *apiAlarm) SendRes(content string) *errmsg.ErrMsg {
	aa.log.Info("IGG SDK res alarm", zap.String("token", aa.cfg.ResToken))
	if aa.cfg.ResToken == "" {
		return nil
	}
	err := aa.aa.Send(aa.cfg.ResToken, content, nil)
	if !err.IsSuccess() {
		return errmsg.NewInternalErr(err.Dump())
	}
	return nil
}
