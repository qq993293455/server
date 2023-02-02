package atlas

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/atlas/dao"
	"coin-server/rule"
	"fmt"
	"strconv"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewAtlasService(
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
	}
	r := rule.MustGetReader(&ctx.Context{}).StoryIllustration.List()
	for _, cnf := range r {
		taskCnf := cnf.UlockCondition
		if len(taskCnf) == 0 {
			continue
		}

		if len(taskCnf) < 3 || len(taskCnf) > 3 {
			panic(fmt.Errorf("StoryIllustration Id %d UlockCondition len %d", cnf.Id, len(taskCnf)))
		}
		s.Module.RegisterCondHandler(models.TaskType(taskCnf[0]), taskCnf[1], taskCnf[2], s.TaskTypHandler, nil)
	}

	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("获取图鉴", s.GetAtlasRequest)
	s.svc.RegisterFunc("解锁图鉴", s.UnlockAtlasRequest)

	eventlocal.SubscribeEventLocal(s.HandleEquipUpdate)
	//eventlocal.SubscribeEventLocal(s.HandleMainTaskUpdate)
	eventlocal.SubscribeEventLocal(s.HandleRelicsUpdate)
	eventlocal.SubscribeEventLocal(s.HandleRoleLoginEvent)
}

func (s *Service) TaskTypHandler(ctx *ctx.Context, d *event.TargetUpdate, _ any) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx).StoryIllustration.List()
	for _, cnf := range r {
		taskCnf := cnf.UlockCondition
		if len(taskCnf) != 3 {
			continue
		}
		if taskCnf[0] == int64(d.Typ) && taskCnf[1] == d.Id && taskCnf[2] <= d.Count {
			err := s.addToAtlas(ctx, models.AtlasType_PictureAtlas, cnf.Id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) HandleRoleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx).StoryIllustration.List()
	for _, cnf := range r {
		taskCnf := cnf.UlockCondition
		if len(taskCnf) == 0 {
			err := s.addToAtlas(ctx, models.AtlasType_PictureAtlas, cnf.Id)
			if err != nil {
				s.log.Error("addToAtlas faild", zap.Any("storyIllustrationId", cnf.Id), zap.Any("role", ctx.RoleId), zap.Any("err", err))
				return err
			}
			continue
		}

		if len(taskCnf) != 3 {
			s.log.Error("StoryIllustration config error", zap.Any("storyIllustrationId", cnf.Id), zap.Any("taskCnf len", len(taskCnf)))
			continue
		}

		cfg, ok := rule.MustGetReader(nil).TaskType.GetTaskTypeById(taskCnf[0])
		if !ok {
			panic(errmsg.NewInternalErr("task_type not found: " + strconv.Itoa(int(taskCnf[0]))))
		}
		unlock := false

		if cfg.IsAccumulate {
			counter, err := s.TaskService.GetCounterByType(ctx, models.TaskType(taskCnf[0]))
			if err != nil {
				panic(err)
			}
			count := counter[taskCnf[1]]
			unlock = count >= taskCnf[2]
		}

		if unlock {
			err := s.addToAtlas(ctx, models.AtlasType_PictureAtlas, cnf.Id)
			if err != nil {
				s.log.Error("addToAtlas faild", zap.Any("storyIllustrationId", cnf.Id), zap.Any("role", ctx.RoleId), zap.Any("err", err))
				return err
			}
		}
	}
	return nil
}

func (s *Service) GetAtlasRequest(ctx *ctx.Context, _ *servicepb.Atlas_GetAtlasRequest) (*servicepb.Atlas_GetAtlasResponse, *errmsg.ErrMsg) {
	atlas, err := dao.GetAtlas(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.Atlas_GetAtlasResponse{Atlas: AtlasDao2Models(atlas)}, nil
}

func (s *Service) UnlockAtlasRequest(ctx *ctx.Context, req *servicepb.Atlas_UnlockAtlasRequest) (*servicepb.Atlas_UnlockAtlasResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	atlasCfg, ok := r.Atlas.AtlasByType(values.Integer(req.Typ))
	if !ok {
		return nil, nil
	}
	cfg, ok := atlasCfg[req.Id]
	if !ok {
		return nil, nil
	}

	atlas, err := dao.GetAtlasByType(ctx, ctx.RoleId, req.Typ)
	if err != nil {
		return nil, err
	}
	if atlas == nil {
		return nil, nil
	}
	if atlas.Each[req.Id] != models.RewardStatus_Unlocked {
		return nil, nil
	}
	atlas.Progress++
	atlas.Each[req.Id] = models.RewardStatus_Received
	// 加到总buff
	if len(cfg.Attrnum) != 0 || len(cfg.Attrrate) != 0 {
		ctx.PublishEventLocal(&event.AttrUpdateToRole{
			Typ:         models.AttrBonusType_TypeAtlas,
			AttrFixed:   cfg.Attrnum,
			AttrPercent: cfg.Attrrate,
		})
	}

	// 获得奖励
	if len(cfg.ItemReward) != 0 {
		_, err = s.BagService.AddManyItem(ctx, ctx.RoleId, cfg.ItemReward)
		if err != nil {
			return nil, err
		}
	}
	dao.SaveAtlas(ctx, ctx.RoleId, atlas)
	return &servicepb.Atlas_UnlockAtlasResponse{
		Typ:    atlas.AtlasType,
		Id:     req.Id,
		Status: atlas.Each[req.Id],
		Reward: cfg.ItemReward,
	}, nil
}

// 只保存激活完毕的图鉴项
func (s *Service) createAtlas(ctx *ctx.Context, typ models.AtlasType) *pbdao.Atlas {
	r := rule.MustGetReader(ctx)

	// 找各类图鉴的ID
	var atlas *pbdao.Atlas
	for idx := range r.Atlas.List() {
		if models.AtlasType(r.Atlas.List()[idx].Id) == typ {
			atlas = &pbdao.Atlas{
				AtlasType: typ,
				Each:      map[int64]models.RewardStatus{},
				Progress:  0,
			}
		}
	}
	return atlas
}

func (s *Service) addToAtlas(ctx *ctx.Context, typ models.AtlasType, id values.Integer) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	atlasByType, ok := r.Atlas.AtlasByType(values.Integer(typ))
	if !ok {
		return nil
	}
	atlas, err := dao.GetAtlasByType(ctx, ctx.RoleId, typ)
	if err != nil {
		return err
	}
	if atlas == nil {
		atlas = s.createAtlas(ctx, typ)
	}
	_, ok = atlasByType[id]
	if !ok {
		return nil
	}
	if _, has := atlas.Each[id]; has {
		return nil
	}
	atlas.Each[id] = models.RewardStatus_Unlocked
	dao.SaveAtlas(ctx, ctx.RoleId, atlas)
	return nil
}

func (s *Service) addMultiToAtlas(ctx *ctx.Context, typ models.AtlasType, ids []values.Integer) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	atlasByType, ok := r.Atlas.AtlasByType(values.Integer(typ))
	if !ok {
		return nil
	}
	atlas, err := dao.GetAtlasByType(ctx, ctx.RoleId, typ)
	if err != nil {
		return err
	}
	if atlas == nil {
		atlas = s.createAtlas(ctx, typ)
	}
	find := false
	for _, id := range ids {
		_, ok = atlasByType[id]
		if ok {
			if _, has := atlas.Each[id]; has {
				continue
			}
			atlas.Each[id] = models.RewardStatus_Unlocked
			find = true
		}
	}
	if find {
		dao.SaveAtlas(ctx, ctx.RoleId, atlas)
	}
	return nil
}
