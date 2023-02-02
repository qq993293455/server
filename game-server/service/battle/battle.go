package battle

import (
	"fmt"

	"coin-server/game-server/service/user/db"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/im"
	"coin-server/common/logger"
	"coin-server/common/proto/cppbattle"
	"coin-server/common/proto/gatewaytcp"
	"coin-server/common/proto/models"
	newcenterpb "coin-server/common/proto/newcenter"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/utils/imutil"
	"coin-server/common/values"
	"coin-server/common/values/env"
	"coin-server/game-server/module"
	"coin-server/game-server/util/trans"
	"coin-server/rule"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	log *logger.Logger,
	module *module.Module,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		log:        log,
		Module:     module,
	}
	s.Module.BattleService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("检查是否可以切换挂机地图", this_.CanToggleBattle)
	this_.svc.RegisterFunc("CPP玩家进入地图", this_.CPPEnterBattle)
	this_.svc.RegisterFunc("获取挂机地图信息", this_.CPPGetBattleServerInfo)
	this_.svc.RegisterFunc("获取玩家当前在哪个地图", this_.GetCurrBattleInfo)
	this_.svc.RegisterFunc("获取所有挂机地图分线", this_.GetMapAllLines)
	this_.svc.RegisterFunc("获取临时背包", this_.GetTempBag)
	this_.svc.RegisterFunc("领取临时背包", this_.DrawTempBag)
	this_.svc.RegisterFunc("客户端战斗自动吃药", this_.AutoMedicineFromClient)

	this_.svc.RegisterFunc("作弊器增加Buff", this_.CheatAddBuff)

	this_.svc.RegisterEvent("挂机地图玩家离线通知", this_.CPPRoleOfflinePush)
	this_.svc.RegisterEvent("CPP吃药事件", this_.CPPAutoMedicine)
	this_.svc.RegisterEvent("临时背包同步事件", this_.HandleTempBagSyncEvent)
	this_.svc.RegisterEvent("同步临时背包", this_.SyncTempBag)
	this_.svc.RegisterEvent("复活扣钻石事件", this_.CppRevive)

	eventlocal.SubscribeEventLocal(this_.HandleRoleLvUpEvent)
	eventlocal.SubscribeEventLocal(this_.HandleMainTaskFinished)
	eventlocal.SubscribeEventLocal(this_.HandleExtraSkillTypTotal)
	eventlocal.SubscribeEventLocal(this_.HandleLoginEvent)
	eventlocal.SubscribeEventLocal(this_.HandleRoleTitleEvent)
}

func (this_ *Service) CPPGetBattleServerInfo(c *ctx.Context, req *servicepb.GameBattle_CPPGetBattleServerInfoRequest) (*servicepb.GameBattle_CPPGetBattleServerInfoResponse, *errmsg.ErrMsg) {
	u, err := this_.Module.GetUserById(c, c.UserId)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errmsg.NewErrUserNotFound()
	}

	if u.HangupMapId == 0 {
		defaultBattleMapID, ok := rule.MustGetReader(c).KeyValue.GetInt64("DefaultBattleMapID")
		if !ok {
			return nil, errmsg.NewErrNoDefaultMap()
		}
		u.HangupMapId = defaultBattleMapID

		out := &newcenterpb.NewCenter_CurBattleInfoResponse{}
		centerId := env.GetCenterServerId()
		err = this_.svc.GetNatsClient().RequestWithOut(c, centerId, &newcenterpb.NewCenter_CurBattleInfoRequest{
			MapId:          u.MapId,
			BattleServerId: u.BattleServerId,
			HungUpMapId:    u.HangupMapId,
			HungUpServerId: u.HangupBattleId,
		}, out)
		if err != nil {
			return nil, err
		}
		if out.LineInfo == nil {
			return nil, errmsg.NewInternalErr("can not find valid hangup line server")
		}
		if out.LineInfo.BattleServerId <= 0 {
			return nil, errmsg.NewInternalErr("can not find valid hangup line server")
		}
		u.HangupBattleId = out.LineInfo.BattleServerId
		this_.Module.SaveUser(c, u)
	}
	return &servicepb.GameBattle_CPPGetBattleServerInfoResponse{
		HangupBattleId: u.HangupBattleId,
		HangupMapId:    u.HangupMapId,
	}, nil
}

func (this_ *Service) CPPRoleOfflinePush(c *ctx.Context, req *servicepb.GameBattle_CPPRoleOfflinePush) {
	u, err := this_.Module.GetUserById(c, c.UserId)
	if err != nil {
		return
	}
	if u == nil {
		return
	}
	r := rule.MustGetReader(c)
	scene, ok := r.MapScene.GetMapSceneById(c.BattleMapId)
	if ok {
		// 离开挂机分线聊天频道
		if scene.MapType == HangUp {
			this_.log.Debug("LeaveRoom", zap.String("roleId", c.RoleId), zap.String("req", req.GoString()))
			if errIm := im.DefaultClient.LeaveRoom(c, &im.RoomRole{
				RoomID:  fmt.Sprintf("battle_line_%d", req.HangupBattleId),
				RoleIDs: []string{c.RoleId},
			}); errIm != nil {
				this_.log.Warn("LeaveRoom fail", zap.Error(err))
			}
		}

		if scene.MapType == HangUp && u.HangupBattleId != req.HangupBattleId {
			return
		}
		if scene.MapType != HangUp {
			u.MapId = u.HangupMapId
			u.BattleServerId = u.HangupBattleId
			this_.Module.SaveUser(c, u)
			return
		}
	}

	u.HangupMapId = req.HangupMapId
	u.HangupBattleId = req.HangupBattleId
	u.HangupPosX = req.HangupPosX
	u.HangupPosY = req.HangupPosY
	this_.Module.SaveUser(c, u)
}

func (this_ *Service) GetStageBattleServerId(c *ctx.Context, sceneId values.Integer) (*models.LineInfo, *errmsg.ErrMsg) {
	ok, err := this_.Module.StageService.IsLock(c, sceneId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrStageNotUnlock()
	}
	r := rule.MustGetReader(c)
	scene, findOk := r.MapScene.GetMapSceneById(sceneId)
	if !findOk || scene.MapType != HangUp {
		return nil, errmsg.NewInternalErr("invalid hungUp mapId")
	}
	u, err := this_.UserService.GetUserById(c, c.UserId)
	if err != nil {
		return nil, err
	}

	battleSrvId := values.Integer(0)
	if u.HangupMapId == sceneId {
		battleSrvId = u.HangupBattleId
	}
	out := &newcenterpb.NewCenter_GetTargetLineResponse{}
	centerId := env.GetCenterServerId()
	err = this_.svc.GetNatsClient().RequestWithOut(c, centerId, &newcenterpb.NewCenter_GetTargetLineRequest{
		MapId:          sceneId,
		BattleServerId: battleSrvId,
	}, out)
	if err != nil {
		return nil, err
	}
	this_.log.Debug("centerResp", zap.String("out", out.GoString()))
	if out.LineInfo == nil || out.LineInfo.BattleServerId <= 0 {
		return nil, errmsg.NewInternalErr("can not find valid hangup line server")
	}
	return out.LineInfo, nil
}

func (this_ *Service) NicknameChange(c *ctx.Context, nickName string) *errmsg.ErrMsg {
	u, err := this_.Module.GetUserById(c, c.UserId)
	if err != nil {
		return err
	}
	if u == nil {
		return errmsg.NewErrUserNotFound()
	}
	return this_.svc.GetNatsClient().PublishCtx(c, u.BattleServerId, &cppbattle.CPPBattle_SyncNicknameChangePush{RoleId: c.RoleId, NickName: nickName})
}

func (this_ *Service) CanToggleBattle(c *ctx.Context, req *servicepb.GameBattle_CPPCanToggleHangupRequest) (*servicepb.GameBattle_CPPCanToggleHangupResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(c)
	newScene, ok := r.MapScene.GetMapSceneById(req.MapId)
	if !ok {
		return nil, errmsg.NewErrMapNotExist()
	}
	oldScene, findOldOk := r.MapScene.GetMapSceneById(c.BattleMapId)
	if c.BattleServerId != 0 && c.BattleMapId != 0 /*&& c.BattleServerId != req.BattleServerId*/ { // 注释的判断是为了兼容客户端，客户端请求这个接口的时候取不到要进的地图的battleId
		if findOldOk && newScene.MapType == HangUp && oldScene.MapType == HangUp { // 都是挂机地图，要去看看玩家死没死
			out := &cppbattle.CPPBattle_CanChangeSceneResponse{}
			err := this_.svc.GetNatsClient().RequestWithOut(c, c.BattleServerId, &cppbattle.CPPBattle_CanChangeSceneRequest{RoleId: c.RoleId}, out)
			if err != nil {
				return nil, err
			}
			switch out.CanStatus {
			case 0: // 先写个switch ，后面可能会有多种情况
			default:
				return nil, errmsg.NewErrCanNotToggle()
			}
		}
	}
	return &servicepb.GameBattle_CPPCanToggleHangupResponse{}, nil
}

func (this_ *Service) CPPEnterBattle(c *ctx.Context, req *servicepb.GameBattle_CPPEnterBattleRequest) (*servicepb.GameBattle_CPPEnterBattleResponse, *errmsg.ErrMsg) {
	// -10000为突破战斗，客户端只需要返回BattleServerId即可
	if req.BattleServerId == -10000 {
		return &servicepb.GameBattle_CPPEnterBattleResponse{
			BattleServerId: req.BattleServerId,
		}, nil
	}
	if req.Pos == nil {
		req.Pos = &cppbattle.CPPBattle_Vec2{}
	}
	if req.Towards == nil {
		req.Towards = &cppbattle.CPPBattle_Vec2{}
	}
	r := rule.MustGetReader(c)
	newScene, ok := r.MapScene.GetMapSceneById(req.MapId)
	if !ok {
		return nil, errmsg.NewErrMapNotExist()
	}

	oldBattleServerId := c.BattleServerId

	u, err := this_.Module.GetUserById(c, c.UserId)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errmsg.NewErrUserNotFound()
	}

	if newScene.MapType == HangUp {
		if req.BattleServerId == u.BattleServerId && req.MapId == u.MapId {
			if req.Pos.X == 0 && req.Pos.Y == 0 {
				req.Pos.X = u.HangupPosX
				req.Pos.Y = u.HangupPosY
			}
		} else { // 如果是挂机地图，发现地图不一样一定要清理原来的坐标点
			u.HangupPosX = 0
			u.HangupPosY = 0
		}
	}

	role, err := this_.Module.GetRoleModelByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	var heroes []*models.Hero

	if req.IsSingle {
		heroCfg, ok := rule.MustGetReader(c).RowHero.GetRowHeroById(req.ConfigId)
		if !ok {
			return nil, errmsg.NewErrHeroNotFound()
		}
		hero, ok, err := this_.Module.GetHero(c, c.RoleId, heroCfg.OriginId)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errmsg.NewErrHeroNotFound()
		}
		heroes = append(heroes, hero)
	} else {
		heroIds := make([]int64, 0, 2)
		heroesFormation, err := this_.Module.FormationService.GetDefaultHeroes(c, c.RoleId)
		if err != nil {
			return nil, err
		}
		if heroesFormation.Hero_0 > 0 && heroesFormation.HeroOrigin_0 > 0 {
			heroIds = append(heroIds, heroesFormation.HeroOrigin_0)
		}
		if heroesFormation.Hero_1 > 0 && heroesFormation.HeroOrigin_1 > 0 {
			heroIds = append(heroIds, heroesFormation.HeroOrigin_1)
			if req.ConfigId == heroesFormation.Hero_1 {
				heroIds[0], heroIds[1] = heroIds[1], heroIds[0]
			}
		}
		heroes, err = this_.Module.GetHeroes(c, c.RoleId, heroIds)
		if err != nil {
			return nil, err
		}
	}
	equips, err := this_.GetManyEquipBagMap(c, c.RoleId, this_.GetHeroesEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	cppHeroes := trans.Heroes2CppHeroes(c, heroes, equips)
	if len(cppHeroes) == 0 {
		return nil, errmsg.NewErrHeroNotFound()
	}

	for _, h := range cppHeroes {
		if len(h.SkillIds) == 0 {
			this_.log.Warn("cppHeros", zap.Any("cppHeroes", cppHeroes))
			return nil, errmsg.NewInternalErr("SkillIds empty")
		}
		if h.Fashion <= 0 {
			this_.log.Warn("cppHeros", zap.Any("cppHeroes", cppHeroes))
			return nil, errmsg.NewInternalErr("fashion empty")
		}
	}

	medicines, err := this_.BagService.GetMedicineMsg(c, c.RoleId, req.MapId)
	if err != nil {
		return nil, err
	}

	d, err00 := db.GetBattleSetting(c)
	if err00 != nil {
		return nil, err00
	}

	// 进入新的地图
	out1 := &cppbattle.CPPBattle_EnterBattleResponse{}
	this_.log.Debug("cpp_enter heroes", zap.Any("heroes", cppHeroes))
	err = this_.svc.GetNatsClient().RequestWithOut(c, req.BattleServerId, &cppbattle.CPPBattle_EnterBattleRequest{
		Pos: &cppbattle.CPPBattle_Vec2{
			X: req.Pos.X,
			Y: req.Pos.Y,
		},
		MapId:          req.MapId,
		BattleServerId: req.BattleServerId,
		Role:           role,
		Heroes:         cppHeroes,
		Medicine:       medicines,
		Towards: &cppbattle.CPPBattle_Vec2{
			X: req.Towards.X,
			Y: req.Towards.Y,
		},
		AutoSouleSkill: d.Data.AutoSoulSkill,
	}, out1)
	if err != nil {
		this_.log.Warn("enter fail", zap.Error(err))
		if err.ErrMsg == "ErrBossHallDead" {
			return &servicepb.GameBattle_CPPEnterBattleResponse{ErrCode: int64(servicepb.BattleErrorCode_ErrBossHallDead)}, nil
		}
		return &servicepb.GameBattle_CPPEnterBattleResponse{ErrCode: int64(servicepb.BattleErrorCode_ErrBattleLineFull)}, nil
	}
	if out1.Pos == nil {
		out1.Pos = &cppbattle.CPPBattle_Vec2{}
	}

	leaveResp := &cppbattle.CPPBattle_LeaveAreaResponse{}
	oldScene, findOldOk := r.MapScene.GetMapSceneById(c.BattleMapId)
	if oldBattleServerId != 0 && oldBattleServerId != req.BattleServerId && findOldOk && oldScene.MapType == HangUp {
		err = this_.svc.GetNatsClient().RequestWithOut(c, oldBattleServerId, &cppbattle.CPPBattle_LeaveAreaRequest{}, leaveResp)
		if err != nil {
			if !err.IsErrorNatsNoResponders() {
				return nil, err
			} else {
				this_.log.Warn("battle not found", zap.Int64("battleServerId", oldBattleServerId), zap.Error(err))
				err = nil
			}
		}
		// 离开挂机分线聊天频道
		if errIm := im.DefaultClient.LeaveRoom(c, &im.RoomRole{
			RoomID:  fmt.Sprintf("battle_line_%d", req.BattleServerId),
			RoleIDs: []string{c.RoleId},
		}); errIm != nil {
			this_.log.Warn("LeaveRoom fail", zap.Error(err))
			return nil, errmsg.NewInternalErr(errIm.Error())
		}
	}

	// 加入挂机分线聊天频道
	if newScene.MapType == HangUp {
		if errIm := im.DefaultClient.JoinRoom(c, &im.RoomRole{
			RoomID:  imutil.BattleLineRoom(req.BattleServerId),
			RoleIDs: []string{c.RoleId},
		}); errIm != nil {
			this_.log.Warn("JoinRoom fail", zap.Error(errIm))
			return nil, errmsg.NewInternalErr(errIm.Error())
		}
	}

	// 告诉网管所在战斗服
	err = this_.svc.GetNatsClient().Publish(c.GateId, c.ServerHeader, &gatewaytcp.GatewayStdTcp_UserChangeBattleId{
		MapId:          req.MapId,
		BattleServerId: req.BattleServerId,
	})
	if err != nil {
		return nil, err
	}
	if newScene.MapType == HangUp {
		u.HangupBattleId = req.BattleServerId
		u.HangupMapId = req.MapId
	}
	u.BattleServerId = req.BattleServerId
	u.MapId = req.MapId

	this_.Module.SaveUser(c, u)
	this_.Module.UpdateTarget(c, c.RoleId, models.TaskType_TaskEnterMapScene, req.MapId, 1)
	return &servicepb.GameBattle_CPPEnterBattleResponse{
		MapId:          req.MapId,
		Ip:             out1.Ip,
		Token:          out1.Token,
		Port:           out1.Port,
		BattleServerId: req.BattleServerId,
	}, nil
}

func (this_ *Service) CppRevive(c *ctx.Context, req *servicepb.Bag_SubItemEvent) {
	this_.Module.SubItem(c, c.RoleId, req.ItemId, req.Count)
}

func (this_ *Service) CPPAutoMedicine(c *ctx.Context, req *servicepb.GameBattle_AutoMedicinePush) {
	err := this_.BagService.AutoTakeMedicine(c, c.RoleId, req.Typ, req.MapId)
	if err != nil {
		this_.log.Warn("CPPAutoMedicine failed!", zap.String("error: ", err.Error()))
	}
	return
}

func (this_ *Service) AutoMedicineFromClient(c *ctx.Context, req *servicepb.GameBattle_AutoMedicineClientRequest) (*servicepb.GameBattle_AutoMedicineClientResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(c)
	cnf, has := r.MapScene.GetMapSceneById(req.MapId)
	if !has || len(cnf.MedicamentId) == 0 {
		return nil, errmsg.NewInternalErr("invalid map id, no medicine")
	}
	err := this_.BagService.AutoTakeMedicine(c, c.RoleId, req.Typ, req.MapId)
	if err != nil {
		this_.log.Warn("CPPAutoMedicine failed!", zap.String("error: ", err.Error()))
	}
	res := &servicepb.GameBattle_AutoMedicineClientResponse{
		MapId:    req.MapId,
		BattleId: req.BattleId,
	}
	val := c.GetValue("CTX_MEDICINE_REQ")
	if msg, ok := val.(*cppbattle.CPPBattle_TakeMedicineRequest); ok {
		res.Id = msg.Id
		res.CdTime = msg.CdTime
	}
	medicines, err1 := this_.BagService.GetMedicineMsg(c, c.RoleId, req.MapId)
	if err != nil {
		return nil, err1
	}
	res.Medicine = medicines
	return res, nil
}

func (this_ *Service) GetMapAllLines(c *ctx.Context, req *servicepb.GameBattle_GetMapAllLinesRequest) (*servicepb.GameBattle_GetMapAllLinesResponse, *errmsg.ErrMsg) {
	u, err := this_.UserService.GetUserById(c, c.UserId)
	if err != nil {
		return nil, err
	}
	res := &servicepb.GameBattle_GetMapAllLinesResponse{}
	if u.HangupMapId != 0 {
		out := &newcenterpb.NewCenter_GetMapAllLinesResponse{}
		centerId := env.GetCenterServerId()
		err = this_.svc.GetNatsClient().RequestWithOut(c, centerId, &newcenterpb.NewCenter_GetMapAllLinesRequest{
			MapId: u.HangupMapId,
		}, out)
		if err != nil {
			return nil, err
		}
		res.AllLineInfo = out.AllLineInfo
	}
	return res, nil
}

func (this_ *Service) CheatAddBuff(c *ctx.Context, req *servicepb.GameBattle_CheatAddBuffRequest) (*servicepb.GameBattle_CheatAddBuffResponse, *errmsg.ErrMsg) {
	u, err := this_.UserService.GetUserById(c, c.UserId)
	if err != nil {
		return nil, err
	}
	err = this_.svc.GetNatsClient().PublishCtx(c, u.BattleServerId, &cppbattle.CPPBattle_GMAddBuffPush{
		BuffId: req.BuffId,
	})
	return &servicepb.GameBattle_CheatAddBuffResponse{}, err
}

const CTX_CUR_BATTLE = "CTX_CUR_BATTLE"

func (this_ *Service) GetCurrBattleInfo(c *ctx.Context, req *servicepb.GameBattle_GetCurrBattleInfoRequest) (*servicepb.GameBattle_GetCurrBattleInfoResponse, *errmsg.ErrMsg) {
	res := &servicepb.GameBattle_GetCurrBattleInfoResponse{}
	if req.SceneId > 0 {
		lineInfo, err := this_.GetStageBattleServerId(c, req.SceneId)
		if err != nil {
			return nil, err
		}
		res.LineInfo = lineInfo
		res.HungupMapId = req.SceneId
		res.BattleId = lineInfo.BattleServerId
		res.SceneId = req.SceneId
		return res, nil
	}

	ctxVal := c.GetValue(CTX_CUR_BATTLE)
	cache, ok := ctxVal.(*servicepb.GameBattle_GetCurrBattleInfoResponse)
	if ok {
		res.LineInfo = cache.LineInfo
		res.HungupMapId = cache.HungupMapId
		res.BattleId = cache.BattleId
		res.SceneId = cache.SceneId
		return res, nil
	}

	u, err := this_.UserService.GetUserById(c, c.UserId)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errmsg.NewInternalErr("useId not exist")
	}
	if u.HangupMapId == 0 {
		return nil, errmsg.NewInternalErr("no default HangupMapId")
	}

	if u.BattleServerId == 0 {
		u.MapId = u.HangupMapId
	}
	out := &newcenterpb.NewCenter_CurBattleInfoResponse{}
	centerId := env.GetCenterServerId()
	err = this_.svc.GetNatsClient().RequestWithOut(c, centerId, &newcenterpb.NewCenter_CurBattleInfoRequest{
		MapId:          u.MapId,
		BattleServerId: u.BattleServerId,
		HungUpMapId:    u.HangupMapId,
		HungUpServerId: u.HangupBattleId,
	}, out)
	if err != nil {
		return nil, err
	}
	c.Debug("center res", zap.Any("out", out))
	if out.BattleServerId != u.BattleServerId || out.HungUpServerId != u.HangupBattleId {
		u.BattleServerId, u.MapId = out.BattleServerId, out.MapId
		u.HangupMapId, u.HangupBattleId = out.HungUpMapId, out.HungUpServerId
		this_.UserService.SaveUser(c, u)
	}
	res.LineInfo = out.LineInfo
	res.HungupMapId = u.HangupMapId
	res.BattleId = u.BattleServerId
	res.SceneId = u.MapId
	c.SetValue(CTX_CUR_BATTLE, res)
	return res, nil
}

func (this_ *Service) GetCurBattleSrvId(c *ctx.Context) (values.Integer, *errmsg.ErrMsg) {
	if c.BattleServerId > 0 {
		return c.BattleServerId, nil
	}
	ctxVal := c.GetValue(CTX_CUR_BATTLE)
	cache, ok := ctxVal.(*servicepb.GameBattle_GetCurrBattleInfoResponse)
	if ok {
		return cache.BattleId, nil
	}
	curRes, err1 := this_.GetCurrBattleInfo(c, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		c.Warn("GetCurrBattleInfo error", zap.Any("err msg", err1))
		return 0, nil
	}
	return curRes.BattleId, nil
}
