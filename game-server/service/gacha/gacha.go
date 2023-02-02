package gacha

import (
	"fmt"
	"math/rand"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/iggsdk"
	"coin-server/common/im"
	"coin-server/common/logger"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/safego"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/common/values/enum/ItemType"
	"coin-server/common/values/enum/Notice"
	event2 "coin-server/common/values/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/gacha/containers"
	"coin-server/game-server/service/gacha/dao"
	"coin-server/rule"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log       *logger.Logger
	lru       containers.LRU
	gachaChan chan values.RoleId
}

func NewGachaService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
		log:        log,
		lru:        containers.NewDefaultLRU(),
		gachaChan:  make(chan values.RoleId, 128),
	}
	safego.GOWithLogger(s.log, func() {
		for roleId := range s.gachaChan {
			cnt := s.lru.Put(roleId)
			if cnt > 1000 {
				logStr := fmt.Sprintf("玩家 %d 1分钟内抽卡超过1000次", utils.Base34DecodeString(roleId))
				s.log.Warn(logStr)
				iggsdk.GetAlarmIns().SendRes(logStr)
				s.lru.Delete(roleId)
			}
		}
	})
	timer.Ticker(1*time.Minute, func() bool {
		s.lru.TTL()
		return true
	})
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("获取卡池", s.GetGachaRequest)
	s.svc.RegisterFunc("抽卡", s.GachaRequest)
	s.svc.RegisterFunc("作弊器解锁所有卡池", s.CheatUnlockGachaRequest)

	eventlocal.SubscribeEventLocal(s.HandleMainTaskFinish)
	//eventlocal.SubscribeEventLocal(s.HandleLevelChange)
	eventlocal.SubscribeEventLocal(s.HandleRoleLoginEvent)

	s.TaskService.RegisterCondHandler(models.TaskType_TaskLevel, 0, event2.AllCount, s.HandleLevelChange, 0)
}

func (s *Service) GetGachaRequest(ctx *ctx.Context, _ *servicepb.Gacha_GetGachaPoolRequest) (*servicepb.Gacha_GetGachaPoolResponse, *errmsg.ErrMsg) {
	gachas, err := dao.GetGacha(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	for _, gacha := range gachas {
		if s.refresh(ctx, GachaDao2Model(gacha)) {
			dao.SaveGacha(ctx, ctx.RoleId, gacha)
		}
	}
	return &servicepb.Gacha_GetGachaPoolResponse{Gacha: GachaDao2Models(gachas)}, nil
}

func (s *Service) GachaRequest(ctx *ctx.Context, req *servicepb.Gacha_GachaRequest) (*servicepb.Gacha_GachaResponse, *errmsg.ErrMsg) {
	gacha, err := dao.GetGachaById(ctx, ctx.RoleId, req.GachaId)
	if err != nil {
		return nil, err
	}
	if gacha == nil {
		return nil, errmsg.NewErrGachaNotExist()
	}
	if s.refresh(ctx, GachaDao2Model(gacha)) {
		dao.SaveGacha(ctx, ctx.RoleId, gacha)
	}

	res, err := s.Gacha(ctx, GachaDao2Model(gacha), req.Count)
	if err != nil {
		return nil, err
	}
	dao.SaveGacha(ctx, ctx.RoleId, gacha)

	s.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGachaNumAcc, 0, req.Count)
	s.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGachaNum, 0, req.Count)
	s.sendNotice(ctx, res)

	s.gachaChan <- ctx.RoleId
	return &servicepb.Gacha_GachaResponse{Items: res, Gacha: GachaDao2Model(gacha)}, nil
}

// 发跑马灯公告
func (s *Service) sendNotice(ctx *ctx.Context, items []*models.Item) {
	var role *daopb.Role
	var err *errmsg.ErrMsg

	for _, item := range items {
		cfg, ok := rule.MustGetReader(ctx).Item.GetItemById(item.ItemId)
		if !ok {
			panic(fmt.Sprintf("item config not found : %d", item.ItemId))
		}
		if cfg.Typ == ItemType.Relics && cfg.Quality > 4 { // 获得橙色以上遗物时发跑马灯公告
			if role == nil {
				role, err = s.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
				if err != nil {
					panic(err)
				}
			}
			err = s.ImService.SendNotice(ctx, im.ParseTypeNoticeRelics, Notice.Relics, role.Nickname, item.ItemId)
			if err != nil {
				s.log.Error("svc.ImService.SendNotice error", zap.Error(err))
			}
		}
	}
}

func (s *Service) CheatUnlockGachaRequest(ctx *ctx.Context, _ *servicepb.Gacha_CheatUnlockGachaRequest) (*servicepb.Gacha_CheatUnlockGachaResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	for idx := range r.Gacha.List() {
		gacha := unlockGacha(ctx, r.Gacha.List()[idx].Id)
		dao.SaveGacha(ctx, ctx.RoleId, GachaModel2Dao(gacha))
	}
	return &servicepb.Gacha_CheatUnlockGachaResponse{}, nil
}

func (s *Service) Gacha(ctx *ctx.Context, gacha *models.Gacha, count values.Integer) ([]*models.Item, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.Gacha.GetGachaById(gacha.GachaId)
	if !ok {
		return nil, errmsg.NewErrGachaNotExist()
	}
	if count != 1 && count != cfg.GachaMultiple {
		return nil, errmsg.NewErrGachaWrongCount()
	}
	if gacha.DailyCount+count > cfg.GachaDailyMax {
		return nil, errmsg.NewErrGachaReachLimit()
	}
	// 检查，免费次数在单抽生效, 否则扣道具
	if gacha.FreeCount != 0 && count == 1 {
		gacha.FreeCount--
	} else {
		itemCount, err := s.BagService.GetItem(ctx, ctx.RoleId, cfg.ItemNeed[0])
		if err != nil {
			return nil, err
		}
		need := cfg.ItemNeed[1] * count
		if itemCount >= need || cfg.Price == -1 {
			err = s.BagService.SubItem(ctx, ctx.RoleId, cfg.ItemNeed[0], need)
			if err != nil {
				return nil, err
			}
		} else {
			err = s.BagService.SubItem(ctx, ctx.RoleId, enum.BoundDiamond, cfg.Price*count)
			if err != nil {
				return nil, err
			}
		}
	}

	res, err := s.drawItem(ctx, gacha, count)
	if err != nil {
		return nil, err
	}
	err = s.BagService.AddManyItemPb(ctx, ctx.RoleId, res...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Service) drawItem(ctx *ctx.Context, gacha *models.Gacha, count values.Integer) ([]*models.Item, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, _ := r.Gacha.GetGachaById(gacha.GachaId)
	gachaWeight, _ := r.Gacha.GachaWeightById(gacha.GachaId)

	// 抽取
	itemCase := make([]values.Integer, 0)
	items := make([]*models.Item, 0)
	for i := 0; i < int(count); i++ {
		gacha.TotalCount++
		gacha.DailyCount++
		// 先判断潜规则抽卡
		itemId, ok := cfg.FixedItem[gacha.TotalCount]
		if ok {
			items = append(items, &models.Item{ItemId: itemId, Count: 1})
			continue
		}
		// 判断保底抽卡
		if gacha.TotalCount%cfg.FixCount == 0 {
			itemCase = append(itemCase, cfg.FixBox)
			continue
		}
		// 根据权重抽卡
		rw := rand.Int63n(gachaWeight.TotalWeight)
		for j := 0; j < len(gachaWeight.GachaIdx); j++ {
			if gachaWeight.GachaWeights[j] >= rw {
				itemCase = append(itemCase, gachaWeight.GachaIdx[j])
				break
			}
		}
	}
	for _, v := range itemCase {
		res, err := s.BagService.UseItemCase(ctx, v, 1, nil)
		if err != nil {
			return nil, err
		}
		// 只会有一个
		for k1, v1 := range res {
			items = append(items, &models.Item{
				ItemId: k1,
				Count:  v1,
			})
			break
		}
	}
	return items, nil
}

func (s *Service) checkAndUnlockGacha(ctx *ctx.Context) ([]*models.Gacha, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	lvCondition, ok := r.Gacha.GachaUnlockCond(1)
	if !ok {
		return nil, nil
	}

	taskCondition, ok := r.Gacha.GachaUnlockCond(2)
	if !ok {
		return nil, nil
	}

	role, err := s.GetRoleModelByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	var gachaSlice []*models.Gacha
	// level
	for lv, gachas := range lvCondition {
		if lv <= role.Level {
			for _, gachaId := range gachas {
				ga, err := dao.GetGachaById(ctx, ctx.RoleId, gachaId)
				if err != nil {
					return nil, err
				}
				if ga == nil {
					hasGacha := false
					for _, gachaData := range gachaSlice {
						if gachaData.GachaId == gachaId {
							hasGacha = true
							break
						}
					}
					if hasGacha {
						continue
					}
					gachaSlice = append(gachaSlice, unlockGacha(ctx, gachaId))
				}
			}
		}
	}
	// task
	for taskId, gachas := range taskCondition {
		isPass, err := s.IsFinishMainTask(ctx, taskId)
		if err != nil {
			return nil, err
		}
		if isPass {
			for _, gachaId := range gachas {
				ga, err := dao.GetGachaById(ctx, ctx.RoleId, gachaId)
				if err != nil {
					return nil, err
				}
				if ga == nil {
					hasGacha := false
					for _, gachaData := range gachaSlice {
						if gachaData.GachaId == gachaId {
							hasGacha = true
							break
						}
					}
					if hasGacha {
						continue
					}
					gachaSlice = append(gachaSlice, unlockGacha(ctx, gachaId))
				}
			}
		}
	}
	return gachaSlice, nil
}

func (s *Service) checkAndUnlock(ctx *ctx.Context, cond values.Integer, value values.Integer) ([]*models.Gacha, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	condition, ok := r.Gacha.GachaUnlockCond(cond)
	if !ok {
		return nil, nil
	}
	count, ok := condition[value]
	if !ok {
		return nil, nil
	}
	var gacha []*models.Gacha
	// 当前条件id：条件数量能解锁的卡池
	for _, gachaId := range count {
		ga, err := dao.GetGachaById(ctx, ctx.RoleId, gachaId)
		if err != nil {
			return nil, err
		}
		if ga == nil {
			gacha = append(gacha, unlockGacha(ctx, gachaId))
		}
	}
	return gacha, nil
}

func (s *Service) refresh(ctx *ctx.Context, gacha *models.Gacha) bool {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.Gacha.GetGachaById(gacha.GachaId)
	if !ok {
		return false
	}
	// 今天已刷新，且配置表没有更改
	if gacha.RefreshAt >= timer.BeginOfDay(timer.Now()).Unix() &&
		cfg.RefreshTime+timer.BeginOfDay(timer.Now()).Unix() == gacha.RefreshAt {
		return false
	}
	// 当前时间没大于每天刷新的时间点
	if timer.Now().Unix() < cfg.RefreshTime+timer.BeginOfDay(timer.Now()).Unix() {
		return false
	}
	gacha.FreeCount = cfg.FreeCount
	gacha.DailyCount = 0
	gacha.RefreshAt = cfg.RefreshTime + timer.BeginOfDay(timer.Now()).Unix()
	return true
}

func unlockGacha(ctx *ctx.Context, gachaId values.GachaId) *models.Gacha {
	r := rule.MustGetReader(ctx)
	cfg, _ := r.Gacha.GetGachaById(gachaId)
	gacha := &models.Gacha{
		GachaId:    gachaId,
		FreeCount:  cfg.FreeCount,
		DailyCount: 0,
		TotalCount: 0,
		RefreshAt:  cfg.RefreshTime + timer.BeginOfDay(timer.Now()).Unix(),
	}
	// 没到今天刷新时间，设为昨天
	if timer.Now().Unix() < cfg.RefreshTime+timer.BeginOfDay(timer.Now()).Unix() {
		gacha.RefreshAt -= 86400
	}
	return gacha
}
