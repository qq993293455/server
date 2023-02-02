package rank

import (
	"fmt"
	"sort"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/orm"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/proto/rank_service"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/redisclient"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/rank/dao"
	rankval "coin-server/game-server/service/rank/values"

	"github.com/rs/xid"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	memDb rankval.MemDb
}

func NewRankService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
		memDb:      rankval.NewMemRankDb(),
	}
	module.RankService = s
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取百人排行榜", svc.GetTopRank)

	eventlocal.SubscribeEventLocal(svc.HandleMemRankValueChangeEvent)
}

func (svc *Service) GetTopRank(ctx *ctx.Context, req *servicepb.Rank_GetTopRankRequest) (*servicepb.Rank_GetTopRankResponse, *errmsg.ErrMsg) {
	out := &rank_service.RankService_TopRankGetResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &rank_service.RankService_TopRankGetRequest{
		Title: req.Title,
	}, out); err != nil {
		return nil, err
	}
	if len(out.RoleId) == 0 {
		return &servicepb.Rank_GetTopRankResponse{
			Title: req.Title,
		}, nil
	}
	ormIns := ctx.NewOrm()
	roles := make([]orm.RedisInterface, 0, len(out.RoleId))
	for _, v := range out.RoleId {
		roles = append(roles, &pbdao.Role{
			RoleId: v,
		})
	}
	notFound, err := ormIns.MGetPB(redisclient.GetUserRedis(), roles...)
	if err != nil {
		return nil, err
	}
	notFoundMap := make(map[int]struct{}, len(notFound))
	for _, v := range notFound {
		notFoundMap[v] = struct{}{}
	}
	dataMap := make(map[values.RoleId]*models.RankItem, 0)
	guildUserList := make([]orm.RedisInterface, 0)
	for i := 0; i < len(roles); i++ {
		if _, ok := notFoundMap[i]; ok {
			continue
		}
		role, ok := roles[i].(*pbdao.Role)
		if !ok {
			continue
		}
		dataMap[role.RoleId] = &models.RankItem{
			RoleId:      role.RoleId,
			Nickname:    role.Nickname,
			Level:       role.Level,
			AvatarId:    role.AvatarId,
			AvatarFrame: role.AvatarFrame,
			Power:       role.Power,
			GuildId:     "",
			GuildName:   "",
			CreatedAt:   role.CreateTime,
		}
		guildUserList = append(guildUserList, &pbdao.GuildUser{
			RoleId: role.RoleId,
		})
	}
	notFound, err = ormIns.MGetPB(redisclient.GetGuildRedis(), guildUserList...)
	if err != nil {
		return nil, err
	}
	notFoundMap = make(map[int]struct{}, len(notFound))
	for _, v := range notFound {
		notFoundMap[v] = struct{}{}
	}
	guildKeys := make([]orm.RedisInterface, 0)
	for i := 0; i < len(guildUserList); i++ {
		if _, ok := notFoundMap[i]; ok {
			continue
		}
		guildUser, ok := guildUserList[i].(*pbdao.GuildUser)
		if !ok || guildUser.GuildId == "" {
			continue
		}
		dataMap[guildUser.RoleId].GuildId = guildUser.GuildId
		guildKeys = append(guildKeys, &pbdao.Guild{Id: guildUser.GuildId})
	}
	notFound, err = ormIns.MGetPB(redisclient.GetGuildRedis(), guildKeys...)
	if err != nil {
		return nil, err
	}
	notFoundMap = make(map[int]struct{}, len(notFound))
	for _, v := range notFound {
		notFoundMap[v] = struct{}{}
	}
	guildNameMap := make(map[values.GuildId]string, 0)
	for i := 0; i < len(guildKeys); i++ {
		if _, ok := notFoundMap[i]; ok {
			continue
		}
		guild, ok := guildKeys[i].(*pbdao.Guild)
		if !ok || guild == nil {
			continue
		}
		guildNameMap[guild.Id] = guild.Name
	}
	list := make([]*models.RankItem, 0, len(dataMap))
	for _, item := range dataMap {
		if item.GuildId != "" {
			item.GuildName = guildNameMap[item.GuildId]
		}
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Power != list[j].Power {
			return list[i].Power > list[j].Power
		}
		if list[i].CreatedAt != list[j].CreatedAt {
			return list[i].CreatedAt < list[j].CreatedAt
		}
		if list[i].Level != list[j].Level {
			return list[i].Level > list[j].Level
		}
		return list[i].RoleId < list[j].RoleId
	})

	return &servicepb.Rank_GetTopRankResponse{
		List:  list,
		Title: req.Title,
	}, nil
}

// ---------------------------------------------------cmd------------------------------------------------------------//

func (svc *Service) CreateRank(ctx *ctx.Context, rankType enum.RankType) values.RankId {
	rankId := xid.New().String()
	rank := rankval.NewRankInfo(rankId, rankType, []*models.RankValue{})
	svc.memDb.SaveRank(rank)
	dao.CreateRank(ctx, rankId)
	return rankId
}

// 新增和更新元素都用此方法
func (svc *Service) UpdateValue(ctx *ctx.Context, value *models.RankValue) *errmsg.ErrMsg {
	rankAgg := svc.memDb.GetRank(value.RankId)
	if rankAgg == nil {
		return errmsg.NewErrRankIdNotExist()
	}
	rankAgg.UpdateValue(ctx, value)
	return nil
}

func (svc *Service) DeleteValue(ctx *ctx.Context, rankId values.RankId, ownerId values.GuildId) *errmsg.ErrMsg {
	rankAgg := svc.memDb.GetRank(rankId)
	if rankAgg == nil {
		return errmsg.NewErrRankIdNotExist()
	}
	rankAgg.DeleteById(ctx, ownerId)
	return nil
}

func (svc *Service) DeleteRank(ctx *ctx.Context, rankId values.RankId) *errmsg.ErrMsg {
	rankAgg := svc.memDb.GetRank(rankId)
	if rankAgg == nil {
		return errmsg.NewErrRankIdNotExist()
	}
	// cmd.repo.DeleteAllValue(ctx, rankId)
	svc.memDb.DeleteRank(rankId)
	return nil
}

func (svc *Service) ClearRanKInMem(ctx *ctx.Context, rankId values.RankId) {
	rankAgg := svc.memDb.GetRank(rankId)
	if rankAgg == nil {
		return
	}
	rankAgg.ClearAll()
	svc.memDb.DeleteRank(rankId)
}

func (svc *Service) InitRankToMem(ctx *ctx.Context, rankId values.RankId, rankType enum.RankType) *errmsg.ErrMsg {
	valuesArray := dao.GetValuesByRankId(ctx, rankId)
	if valuesArray == nil {
		return errmsg.NewErrRankIdNotExist()
	}
	rankAgg := rankval.NewRankInfo(rankId, rankType, valuesArray)
	svc.memDb.SaveRank(rankAgg)
	return nil
}

func (svc *Service) MemRankValueChange(ctx *ctx.Context, data *event.MemRankValueChangeData) *errmsg.ErrMsg {
	for i := range data.AddList {
		dao.CreateValue(ctx, data.AddList[i])
	}
	for i := range data.UpdateList {
		dao.SaveValue(ctx, data.UpdateList[i])
	}
	for i := range data.DeleteList {
		dao.DeleteValue(ctx, data.DeleteList[i])
	}
	return nil
}

// ---------------------------------------------------query------------------------------------------------------------//

func (svc *Service) GetRank(_ *ctx.Context, rankId values.RankId) (rank rankval.RankAgg) {
	return svc.memDb.GetRank(rankId)
}

func (svc *Service) GetTotalNum(_ *ctx.Context, rankId values.RankId) (values.Integer, *errmsg.ErrMsg) {
	rank := svc.memDb.GetRank(rankId)
	if rank == nil {
		return 0, errmsg.NewErrRankIdNotExist()
	}
	return rank.GetTotalNum(), nil
}

func (svc *Service) GetValueById(_ *ctx.Context, rankId values.RankId, ownerId values.GuildId) (*models.RankValue, *errmsg.ErrMsg) {
	rank := svc.memDb.GetRank(rankId)
	if rank == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	return rank.GetValueById(ownerId), nil
}

func (svc *Service) GetScoreValue1ByIds(ctx *ctx.Context, rankId values.RankId, ownerIds []values.GuildId) (map[values.GuildId]values.Integer, *errmsg.ErrMsg) {
	rank := svc.memDb.GetRank(rankId)
	if rank == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	return rank.GetScoreValue1ByIds(ownerIds), nil
}

func (svc *Service) GetScoreValue2ByIds(ctx *ctx.Context, rankId values.RankId, ownerIds []values.GuildId) (map[values.GuildId]models.RankAndScore, *errmsg.ErrMsg) {
	rank := svc.memDb.GetRank(rankId)
	if rank == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	return rank.GetScoreValue2ByIds(ownerIds), nil
}

func (svc *Service) GetValueByRank(ctx *ctx.Context, rankId values.RankId, rank values.Integer) (*models.RankValue, *errmsg.ErrMsg) {
	rankAgg := svc.memDb.GetRank(rankId)
	if rankAgg == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	return rankAgg.GetValueByRank(rank), nil
}

func (svc *Service) GetValueByRange(ctx *ctx.Context, rankId values.RankId, start, end values.Integer) ([]*models.RankValue, *errmsg.ErrMsg) {
	rank := svc.memDb.GetRank(rankId)
	if rank == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	return rank.GetValueByRange(start, end), nil
}

// ---------------------------------------------------event------------------------------------------------------------//

func (svc *Service) HandleMemRankValueChangeEvent(ctx *ctx.Context, d *event.MemRankValueChangeData) *errmsg.ErrMsg {
	ctx.Info(fmt.Sprintf("Mem rank value update, RankId:%s", d.RankId))
	return svc.MemRankValueChange(ctx, d)
}
