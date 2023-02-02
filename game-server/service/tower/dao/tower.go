package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/timer"
	"coin-server/common/values"
)

func Ternary(bflag bool, first_value, second_value interface{}) interface{} {
	if bflag {
		return first_value
	}
	return second_value
}

func CreateTowerData(roleId values.RoleId, ctype models.TowerType) *dao.Tower {
	ret := &dao.Tower{
		RoleId: roleId,
		Data: []*models.TowerData{
			{
				Type:              ctype,
				CurrentTowerLevel: 1,
				PassData:          map[int64]*models.TowerPassData{},
				CacheInfo: &models.TowerAccumulateHarvestCacheInfo{
					ProfitsData:        []*models.Item{},
					LastSettlementTime: timer.Now().Unix(),
					SettlementUseTime:  0,
				},
				MeditationData: &models.MeditationData{
					LastMeditationTime:     timer.Now().Unix(),
					UseFreeMeditationTimes: 0,
					UseCostMeditationTimes: 0,
				},
				ChallengeData: &models.ChallengeData{
					UseChallengeTimes: 0,
					LastChallengeTime: timer.Now().Unix(),
				},
			},
		},
	}
	return ret
}

func UpdateTowerData(ctx *ctx.Context, tower *dao.Tower) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), tower)
}

func GetTowerData(ctx *ctx.Context, roleId values.RoleId) (TowerI, *errmsg.ErrMsg) {
	ret := &dao.Tower{RoleId: roleId}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		ret = CreateTowerData(roleId, models.TowerType_TT_Default)
		UpdateTowerData(ctx, ret)
	}
	return NewTower(ctx, ret), nil
}

type TowerI interface {
	RoleId() values.RoleId
	RefreshMeditationData(ctx *ctx.Context)
	InitTowerRewardData(ctx *ctx.Context, ctype models.TowerType, modTime int64)
	SaveData(ctx *ctx.Context)
	TowerData() *dao.Tower
	AddMeditationTimes(ctx *ctx.Context, ctype models.TowerType, is_free bool)
	RefreshChallenge(ctx *ctx.Context)
	AddTowerLevel(ctype models.TowerType)
}

func NewTower(ctx *ctx.Context, values *dao.Tower) TowerI {
	return &DaoTowerData{
		dao_tower_data: values,
		//ctx:            ctx,
	}
}

type DaoTowerData struct {
	dao_tower_data *dao.Tower
	//ctx            *ctx.Context
}

func (this_ *DaoTowerData) AddTowerLevel(ctype models.TowerType) {
	for _, tower_data := range this_.dao_tower_data.Data {
		if tower_data.Type == ctype {
			tower_data.ChallengeData.LastChallengeTime = timer.Now().Unix()
			tower_data.ChallengeData.UseChallengeTimes++
			tower_data.CurrentTowerLevel++
		}
	}
}

func (this_ *DaoTowerData) RoleId() values.RoleId {
	return this_.dao_tower_data.RoleId
}

func (this_ *DaoTowerData) TowerData() *dao.Tower {
	return this_.dao_tower_data
}

func (this_ *DaoTowerData) RefreshChallenge(ctx *ctx.Context) {
	for _, towerData := range this_.dao_tower_data.Data {
		towerData.ChallengeData.LastChallengeTime = timer.Now().Unix()
		towerData.ChallengeData.UseChallengeTimes = 0
	}
	UpdateTowerData(ctx, this_.dao_tower_data)
}

func (this_ *DaoTowerData) RefreshMeditationData(ctx *ctx.Context) {
	for _, towerData := range this_.dao_tower_data.Data {
		towerData.MeditationData.LastMeditationTime = timer.Now().Unix()
		towerData.MeditationData.UseFreeMeditationTimes = 0
		towerData.MeditationData.UseCostMeditationTimes = 0
	}
	UpdateTowerData(ctx, this_.dao_tower_data)
}

func (this_ *DaoTowerData) InitTowerRewardData(ctx *ctx.Context, ctype models.TowerType, modTime int64) {
	for _, towerData := range this_.dao_tower_data.Data {
		if towerData.Type == ctype {
			towerData.CacheInfo.LastSettlementTime = timer.Now().Unix() - modTime
			towerData.CacheInfo.ProfitsData = []*models.Item{}
			towerData.CacheInfo.SettlementUseTime = 0
			UpdateTowerData(ctx, this_.dao_tower_data)
			break
		}
	}
}

func (this_ *DaoTowerData) SaveData(ctx *ctx.Context) {
	UpdateTowerData(ctx, this_.dao_tower_data)
}

func (this_ *DaoTowerData) AddMeditationTimes(ctx *ctx.Context, ctype models.TowerType, is_free bool) {
	for _, towerData := range this_.dao_tower_data.Data {
		if towerData.Type == ctype {
			if is_free {
				towerData.MeditationData.UseFreeMeditationTimes++
			} else {
				towerData.MeditationData.UseCostMeditationTimes++
			}
			UpdateTowerData(ctx, this_.dao_tower_data)
			break
		}
	}
}
