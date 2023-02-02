package activity

import (
	"time"

	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	rulemodel "coin-server/rule/rule-model"
)

func NewActivityModel(cfg *rulemodel.Activity) (*models.Activity, *errmsg.ErrMsg) {
	begin, err := time.ParseInLocation("2006-01-02 15:04:05", cfg.ActivityOpenTime, time.UTC)
	if err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	end, err := time.ParseInLocation("2006-01-02 15:04:05", cfg.DurationTime, time.UTC)
	if err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	return &models.Activity{
		Id:                         cfg.Id,
		Sort:                       cfg.ActivitySort,
		ActivityDescribeLanguageId: cfg.ActivityDescribeLanguageId,
		Begin:                      begin.UnixMilli(),
		End:                        end.UnixMilli(),
		SystemId:                   cfg.SystemId,
		ChargeId:                   cfg.ChargeId,
	}, nil
}
