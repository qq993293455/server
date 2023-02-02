package map_event

import (
	"math/rand"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/game-server/service/map-event/dao"
	"coin-server/rule"
)

func (s *Service) TreasureChestRequest(ctx *ctx.Context, req *servicepb.MapGame_TreasureChestRequest) (*servicepb.MapGame_TreasureChestResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(req.EventId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	gameId := int64(0)
	if cfg.Typ == 1 {
		var err *errmsg.ErrMsg
		gameId, err = s.checkGameValid(ctx, TreasureChestID)
		if err != nil {
			return nil, err
		}
	}

	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}

	if cfg.Typ == 2 {
		datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
		if err != nil {
			return nil, err
		}
		for _, data := range datas {
			if data.EventId == req.EventId && !data.IsOver {
				gameId = req.GameId
				break
			}
		}
	}

	if gameId == 0 {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	game, ok := r.Anecdotes.CustomAnecdoteGame1()[gameId]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	option, ok := game[req.Option]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	if !req.SkillSuccess {
		err := s.finishEvent(ctx, gameId, req.EventId, false, map[values.ItemId]values.Integer{}, 0, 1)
		if err != nil {
			return nil, err
		}
		return &servicepb.MapGame_TreasureChestResponse{}, nil
	} else {
		if len(option.Item) > 0 {
			err := s.BagService.SubManyItem(ctx, ctx.RoleId, option.Item)
			if err != nil {
				if err == errmsg.NewErrBagNotEnough() || err == errmsg.NewErrBagNoSuchItem() {
					return &servicepb.MapGame_TreasureChestResponse{}, err
				}
				return nil, err
			}
		}
		gameCfg, _ := r.AnecdotesGame1.GetAnecdotesGame1ById(req.GameId)
		reward := map[values.Integer]values.Integer{}
		for k, v := range gameCfg.Reward {
			reward[k] = v
		}
		var ratio = values.Integer(1)
		if len(option.DropList) > 0 {
			randNum := rand.Int63n(10000)
			currNum := values.Integer(1)
			currTimes := values.Integer(1)
			for times, num := range option.DropList {
				currNum += num
				if randNum < currNum {
					currTimes = times
					ratio = currTimes
					break
				}
			}
			for itemId := range reward {
				reward[itemId] *= currTimes
			}
		}
		err := s.finishEvent(ctx, gameId, req.EventId, true, reward, 0, ratio)
		if err != nil {
			return nil, err
		}
		return &servicepb.MapGame_TreasureChestResponse{}, nil
	}
}

func (s *Service) CollectMineRequest(ctx *ctx.Context, req *servicepb.MapGame_CollectMineRequest) (*servicepb.MapGame_CollectMineResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(req.EventId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	gameId := int64(0)
	if cfg.Typ == 1 {
		var err *errmsg.ErrMsg
		gameId, err = s.checkGameValid(ctx, CollectMineID)
		if err != nil {
			return nil, err
		}
	}

	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}

	if cfg.Typ == 2 {
		datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
		if err != nil {
			return nil, err
		}
		for _, data := range datas {
			if data.EventId == req.EventId && !data.IsOver {
				gameId = req.GameId
				break
			}
		}
	}

	if gameId == 0 {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	gameRule, ok := rule.MustGetReader(ctx).AnecdotesGame2.GetAnecdotesGame2ById(gameId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	collect := req.Collect
	if req.Collect > gameRule.SkillNum {
		collect = gameRule.SkillNum
	}
	reward := make(map[values.Integer]values.Integer)
	for item := range gameRule.Reward {
		reward[item] = collect
		break
	}
	err := s.finishEvent(ctx, gameId, req.EventId, true, reward, gameRule.DropListId, 1)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapGame_CollectMineResponse{}, nil
}

func (s *Service) BlackMarketRequest(ctx *ctx.Context, req *servicepb.MapGame_BlackMarketRequest) (*servicepb.MapGame_BlackMarketResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(req.EventId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	gameId := int64(0)
	if cfg.Typ == 1 {
		var err *errmsg.ErrMsg
		gameId, err = s.checkGameValid(ctx, BlackMarketID)
		if err != nil {
			return nil, err
		}
	}

	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}

	if cfg.Typ == 2 {
		datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
		if err != nil {
			return nil, err
		}
		for _, data := range datas {
			if data.EventId == req.EventId && !data.IsOver {
				gameId = req.GameId
				break
			}
		}
	}

	if gameId == 0 {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	game, ok := r.Anecdotes.CustomAnecdoteGame3()[gameId]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	option, ok := game[req.Option]
	if !ok {
		err := s.finishEvent(ctx, gameId, req.EventId, false, nil, 0, 1)
		if err != nil {
			return nil, err
		}
		return &servicepb.MapGame_BlackMarketResponse{}, nil
	}
	err := s.BagService.SubManyItem(ctx, ctx.RoleId, option.Purchase)
	if err != nil {
		return nil, err
	}

	ids := make([]values.Integer, 0)
	weights := make([]values.Integer, 0)
	weight := int64(0)

	for k, v := range option.EquipReward {
		weight += v
		ids = append(ids, k)
		weights = append(weights, weight)
	}

	rn := rand.Int63n(weight)
	i := 0
	for ; i < len(weights); i++ {
		if rn < weights[i] {
			reward := map[values.ItemId]values.Integer{ids[i]: 1}
			for k, v := range option.Reward {
				reward[k] += v
			}
			err = s.finishEvent(ctx, gameId, req.EventId, true, reward, 0, 1)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	return &servicepb.MapGame_BlackMarketResponse{DrawId: ids[i]}, nil
}

func (s *Service) JigsawPuzzleRequest(ctx *ctx.Context, req *servicepb.MapGame_JigsawPuzzleRequest) (*servicepb.MapGame_JigsawPuzzleResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(req.EventId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	gameId := int64(0)
	if cfg.Typ == 1 {
		var err *errmsg.ErrMsg
		gameId, err = s.checkGameValid(ctx, JigsawPuzzleID)
		if err != nil {
			return nil, err
		}
	}

	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}

	if cfg.Typ == 2 {
		datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
		if err != nil {
			return nil, err
		}
		for _, data := range datas {
			if data.EventId == req.EventId && !data.IsOver {
				gameId = req.GameId
				break
			}
		}
	}

	if gameId == 0 {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	game, ok := r.Anecdotes.CustomAnecdoteGame4()[gameId]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	option, ok := game[req.Opt]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	if req.IsSuccess {
		err := s.finishEvent(ctx, gameId, req.EventId, true, option.Reward, option.DropListId, 1)
		if err != nil {
			return nil, err
		}
	} else {
		err := s.finishEvent(ctx, gameId, req.EventId, false, nil, 0, 1)
		if err != nil {
			return nil, err
		}
	}
	return &servicepb.MapGame_JigsawPuzzleResponse{}, nil
}

func (s *Service) CardMatchRequest(ctx *ctx.Context, req *servicepb.MapGame_CardMatchRequest) (*servicepb.MapGame_CardMatchResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(req.EventId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	gameId := int64(0)
	if cfg.Typ == 1 {
		var err *errmsg.ErrMsg
		gameId, err = s.checkGameValid(ctx, CardMatchID)
		if err != nil {
			return nil, err
		}
	}

	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}

	if cfg.Typ == 2 {
		datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
		if err != nil {
			return nil, err
		}
		for _, data := range datas {
			if data.EventId == req.EventId && !data.IsOver {
				gameId = req.GameId
				break
			}
		}
	}

	if gameId == 0 {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	game, ok := r.Anecdotes.CustomAnecdoteGame5()[gameId]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	option, ok := game[req.Opt]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	if req.IsSuccess {
		err := s.finishEvent(ctx, gameId, req.EventId, true, option.Reward, option.DropListId, 1)
		if err != nil {
			return nil, err
		}
	} else {
		err := s.finishEvent(ctx, gameId, req.EventId, false, nil, 0, 1)
		if err != nil {
			return nil, err
		}
	}
	return &servicepb.MapGame_CardMatchResponse{}, nil
}

func (s *Service) WantingRequest(ctx *ctx.Context, req *servicepb.MapGame_WantingRequest) (*servicepb.MapGame_WantingResponse, *errmsg.ErrMsg) {
	err := s.finishEvent(ctx, WantingID, WantingID, true, nil, 0, 1)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapGame_WantingResponse{}, nil
}

func (s *Service) MeetRequest(ctx *ctx.Context, req *servicepb.MapGame_MeetRequest) (*servicepb.MapGame_MeetResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(req.EventId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	gameId := int64(0)
	if cfg.Typ == 1 {
		var err *errmsg.ErrMsg
		gameId, err = s.checkGameValid(ctx, MeetID)
		if err != nil {
			return nil, err
		}
	}

	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}

	if cfg.Typ == 2 {
		datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
		if err != nil {
			return nil, err
		}
		for _, data := range datas {
			if data.EventId == req.EventId && !data.IsOver {
				gameId = req.GameId
				break
			}
		}
	}

	if gameId == 0 {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	game, ok := r.Anecdotes.CustomAnecdoteGame7()[gameId]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	option, ok := game[req.Opt]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	if !req.IsSucc {
		err := s.finishEvent(ctx, gameId, req.EventId, false, map[values.ItemId]values.Integer{}, 0, 1)
		if err != nil {
			return nil, err
		}
		return &servicepb.MapGame_MeetResponse{}, nil
	} else {
		gameCfg, _ := r.AnecdotesGame7.GetAnecdotesGame7ById(req.GameId)
		reward := map[values.Integer]values.Integer{}
		for k, v := range gameCfg.Reward {
			reward[k] = v
		}
		var ratio = values.Integer(1)
		if len(option.DropList) > 0 {
			randNum := rand.Int63n(10000)
			currNum := values.Integer(1)
			currTimes := values.Integer(1)
			for times, num := range option.DropList {
				currNum += num
				if randNum < currNum {
					currTimes = times
					ratio = currTimes
					break
				}
			}
			for itemId := range reward {
				reward[itemId] *= currTimes
			}
		}
		err := s.finishEvent(ctx, gameId, req.EventId, true, reward, 0, ratio)
		if err != nil {
			return nil, err
		}
	}
	return &servicepb.MapGame_MeetResponse{}, nil
}

func (s *Service) MeetBattleStart(ctx *ctx.Context, req *servicepb.MapGame_MeetBattleStartRequest) (*servicepb.MapGame_MeetBattleStartResponse, *errmsg.ErrMsg) {
	/*gameId, err := s.checkGameValid(ctx, MeetID)
	if err != nil {
		return nil, err
	}
	r := rule.MustGetReader(ctx)
	game, ok := r.Anecdotes.CustomAnecdoteGame7()[gameId]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	option, ok := game[req.Opt]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	mapId := option.MapScene
	tokenInfo := TokenInfo{
		RoleId: ctx.RoleId,
		Option: req.Opt,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		TokenInfo: tokenInfo,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: timer.StartTime(ctx.StartTime).Add(time.Minute * 5).Unix(),
		},
	})
	_token, err1 := token.SignedString(JwtKey)
	if err1 != nil {
		return nil, errmsg.NewInternalErr("jwt sign failed")
	}

	role, err := s.Module.GetRoleModelByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	// 获取英雄信息
	heroesFormation, err := s.Module.FormationService.GetDefaultHeroes(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	heroIds := make([]int64, 0, 2)
	if heroesFormation.Hero_0 > 0 && heroesFormation.HeroOrigin_0 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_0)
	}
	if heroesFormation.Hero_1 > 0 && heroesFormation.HeroOrigin_1 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_1)
	}
	heroes, err := s.Module.GetHeroes(ctx, ctx.RoleId, heroIds)
	if err != nil {
		return nil, err
	}
	equips, err := s.GetManyEquipBagMap(ctx, ctx.RoleId, s.GetHeroesEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	cppHeroes := trans.Heroes2CppHeroes(heroes, equips)
	if len(cppHeroes) == 0 {
		return nil, errmsg.NewErrHeroNotFound()
	}

	return &servicepb.MapGame_MeetBattleStartResponse{
		MapId:    mapId,
		Token:    _token,
		BattleId: -10000, // 直接写死-10000 客户端必须，对于服务器无意义
		Sbp: &models.SingleBattleParam{
			Role:             role,
			Heroes:           cppHeroes,
			MonsterGroupInfo: option.MonsterGroupInfo,
			CountDown:        999,
		},
	}, nil*/
	return nil, nil
}

func (s *Service) MeetBattleFinish(ctx *ctx.Context, req *servicepb.MapGame_MeetBattleFinishRequest) (*servicepb.MapGame_MeetBattleFinishResponse, *errmsg.ErrMsg) {
	/*token, err1 := jwt.ParseWithClaims(req.Token, &Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		return JwtKey, nil
	})
	if err1 != nil {
		return nil, errmsg.NewInternalErr(err1.Error())
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errmsg.NewProtocolErrorInfo("invalid token")
	}
	if claims.RoleId != ctx.RoleId {
		return nil, errmsg.NewProtocolErrorInfo("invalid token")
	}
	r, err := db.GetRole(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	gameId, err := s.checkGameValid(ctx, MeetID)
	if err != nil {
		return nil, err
	}
	ru := rule.MustGetReader(ctx)
	game, ok := ru.Anecdotes.CustomAnecdoteGame7()[gameId]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	option, ok := game[claims.Option]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	if req.IsSuccess {
		err = s.finishEvent(ctx, gameId, MeetID, true, option.Reward, 0)
		if err != nil {
			return nil, err
		}
	} else {
		err = s.finishEvent(ctx, gameId, MeetID, false, nil, 0)
		if err != nil {
			return nil, err
		}
	}*/
	return &servicepb.MapGame_MeetBattleFinishResponse{}, nil
}

func (s *Service) BuildArmsRequest(ctx *ctx.Context, req *servicepb.MapGame_BuildArmsRequest) (*servicepb.MapGame_BuildArmsResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(req.EventId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	gameId := int64(0)
	if cfg.Typ == 1 {
		var err *errmsg.ErrMsg
		gameId, err = s.checkGameValid(ctx, BuildArmsID)
		if err != nil {
			return nil, err
		}
	}

	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}

	if cfg.Typ == 2 {
		datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
		if err != nil {
			return nil, err
		}
		for _, data := range datas {
			if data.EventId == req.EventId && !data.IsOver {
				gameId = req.GameId
				break
			}
		}
	}

	if gameId == 0 {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	game, ok := r.Anecdotes.CustomAnecdoteGame8()[gameId]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	option, ok := game[req.Opt]
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	if req.IsSuccess {
		err := s.finishEvent(ctx, gameId, req.EventId, true, option.Reward, 0, 1)
		if err != nil {
			return nil, err
		}
	} else {
		err := s.finishEvent(ctx, gameId, req.EventId, false, option.FailReward, 0, 1)
		if err != nil {
			return nil, err
		}
	}
	return &servicepb.MapGame_BuildArmsResponse{}, nil
}

func (s *Service) checkGameValid(c *ctx.Context, gameTyp int64) (int64, *errmsg.ErrMsg) {
	e, err := dao.GetMapEvent(c, c.RoleId)
	if err != nil {
		return -1, err
	}
	if e == nil {
		return -1, errmsg.NewErrMapEventLock()
	}
	gameId := int64(0)
	for _, event := range e.Triggered {
		if event.EventId == gameTyp {
			gameId = event.GameId
			break
		}
	}
	if gameId == 0 {
		return -1, errmsg.NewErrMapGameNotExist()
	}
	return gameId, nil
}
