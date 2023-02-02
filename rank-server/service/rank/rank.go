package rank

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/proto/rank_service"
	"coin-server/common/redisclient"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/rank-server/service/rank/dao"
	rankval "coin-server/rank-server/service/rank/values"

	"github.com/go-redis/redis/v8"

	"github.com/rs/xid"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	memDb      rankval.MemDb
	lock       *sync.RWMutex
}

func NewRankService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		log:        log,
		memDb:      rankval.NewMemRankDb(),
		lock:       &sync.RWMutex{},
	}
	return s
}

func (svc *Service) Router() {
	h := svc.svc.Group(handler.GameServerAuth)
	h.RegisterFunc("创建排行榜", svc.Create)
	h.RegisterFunc("更新排行榜", svc.UpdateRankValue)
	h.RegisterFunc("批量更新排行榜", svc.BatchUpdateRankValue)
	h.RegisterFunc("范围获取排行榜数据", svc.GetRankValueByIndex)

	h.RegisterFunc("通过OwnerId删除取排行榜数据", svc.DeleteRankValue)
	svc.svc.RegisterFunc("批量通过OwnerId获取排行榜数据", svc.GetRankByOwnerIdS, handler.GVGServerAuth)
	svc.svc.RegisterFunc("通过OwnerId获取排行榜数据", svc.GetRankByOwnerId, handler.GVGOrGameServerAuth)

	// 百人榜
	h.RegisterFunc("获取百人榜数据", svc.GetTopRank)
	h.RegisterFunc("更新百人榜数据", svc.UpdateTopRank)

	eventlocal.SubscribeEventLocal(svc.HandleMemRankValueChangeEvent)
}

const count = 10000

func (svc *Service) Load2Mem() {
	list, err := redisclient.GetDefaultRedis().MGet(context.Background(), enum.AllRankKey()...).Result()
	if err != nil && err != redis.Nil {
		panic(err)
	}
	for _, id := range list {
		if id == nil {
			continue
		}
		rankId, ok := id.(string)
		if !ok {
			panic(fmt.Errorf("assert rankId error:%v", rankId))
		}
		temp := strings.Split(rankId, "@")
		if len(temp) != 2 {
			panic(fmt.Errorf("rankId format error:%s", rankId))
		}
		rankType, err := strconv.Atoi(temp[0])
		if err != nil {
			panic(fmt.Errorf("rankType format error:%s", rankId))
		}
		rankInfo := rankval.NewRankInfo(rankId, enum.RankType(rankType), []*models.RankValue{})
		svc.memDb.SaveRank(rankInfo)
		svc.load2MemByRankType(enum.RankType(rankType), rankId, 0, count)
	}
}

func (svc *Service) load2MemByRankType(typ enum.RankType, rankId values.RankId, start, end int) {
	daoIns := dao.GetDao()
	list, err := daoIns.Get(typ, start, end)
	if err != nil {
		panic(fmt.Errorf("获取 type=%d 排行榜数据失败:%s", typ, err.Error()))
	}
	if len(list) <= 0 {
		return
	}
	ranAgg := svc.memDb.GetRank(rankId)
	for _, value := range list {
		ranAgg.InitValue(&models.RankValue{
			RankId:  value.RankId,
			OwnerId: value.OwnerId,
			Rank:    value.Rank,
			Value1:  value.Value1,
			Value2:  value.Value2,
			Value3:  value.Value3,
			Value4:  value.Value4,
			Extra:   dao.UnmarshalExtra(value.Extra),
			Shard:   value.Shard,
		})
	}
	start = end
	end = start + count
	svc.load2MemByRankType(typ, rankId, start, end)
}

func (svc *Service) Create(_ *ctx.Context, req *rank_service.RankService_CreateRankRequest) (*rank_service.RankService_CreateRankResponse, *errmsg.ErrMsg) {
	rankType := enum.RankType(req.RankType)
	rankId := req.RankId
	if rankId == "" {
		var err *errmsg.ErrMsg
		rankId, err = svc.genRankId(rankType)
		if err != nil {
			return nil, err
		}
	} else {
		if strings.Index(rankId, "@") == -1 {
			rankId = strconv.Itoa(int(rankType)) + "@" + rankId
		} else {
			temp := strings.Split(rankId, "@")
			if len(temp) != 2 {
				return nil, errmsg.NewProtocolErrorInfo("invalid rankId")
			}
			// TODO 检查rankType是否合法
		}
	}

	rank := rankval.NewRankInfo(rankId, rankType, []*models.RankValue{})
	svc.memDb.SaveRank(rank)
	return &rank_service.RankService_CreateRankResponse{
		RankId: rankId,
	}, nil
}

func (svc *Service) UpdateRankValue(ctx *ctx.Context, req *rank_service.RankService_UpdateRankValueRequest) (*rank_service.RankService_UpdateRankValueResponse, *errmsg.ErrMsg) {
	value := req.RankValue
	if value == nil {
		return nil, errmsg.NewProtocolErrorInfo("rank value is nil")
	}
	rankAgg := svc.memDb.GetRank(value.RankId)
	if rankAgg == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	rankAgg.UpdateValue(ctx, value)
	return &rank_service.RankService_UpdateRankValueResponse{}, nil
}

func (svc *Service) BatchUpdateRankValue(ctx *ctx.Context, req *rank_service.RankService_BatchUpdateRankValueRequest) (*rank_service.RankService_BatchUpdateRankValueResponse, *errmsg.ErrMsg) {
	list := req.RankValue
	if len(list) == 0 {
		return nil, errmsg.NewProtocolErrorInfo("rank value is nil")
	}

	rankAgg := svc.memDb.GetRank(list[0].RankId)
	if rankAgg == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	for _, value := range list {
		rankAgg.UpdateValue(ctx, value)
	}

	return &rank_service.RankService_BatchUpdateRankValueResponse{}, nil
}

func (svc *Service) GetRankValueByIndex(_ *ctx.Context, req *rank_service.RankService_GetValueByIndexRequest) (*rank_service.RankService_GetValueByIndexResponse, *errmsg.ErrMsg) {
	rank := svc.memDb.GetRank(req.RankId)
	if rank == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	total := rank.GetTotalNum()
	start := req.Start // start最小从1开始
	if start < 1 {
		start = 1
	}
	end := req.End
	if end > total {
		end = total
	}
	value := rank.GetValueByRange(start, end)
	ending := end >= total
	return &rank_service.RankService_GetValueByIndexResponse{
		RankValues: value,
		Ending:     ending,
	}, nil
}

func (svc *Service) GetRankByOwnerId(_ *ctx.Context, req *rank_service.RankService_GetRankValueByOwnerIdRequest) (*rank_service.RankService_GetRankValueByOwnerIdResponse, *errmsg.ErrMsg) {
	rank := svc.memDb.GetRank(req.RankId)
	if rank == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	val := rank.GetValueById(req.OwnerId)
	return &rank_service.RankService_GetRankValueByOwnerIdResponse{
		RankValue: val,
	}, nil
}

func (svc *Service) GetRankByOwnerIdS(_ *ctx.Context, req *rank_service.RankService_GetRankValueSByOwnerIdSRequest) (*rank_service.RankService_GetRankValueSByOwnerIdSResponse, *errmsg.ErrMsg) {
	rank := svc.memDb.GetRank(req.RankId)
	if rank == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}
	out := map[string]*models.RankValue{}
	for _, v := range req.OwnerIds {
		val := rank.GetValueById(v)
		out[v] = val
	}
	return &rank_service.RankService_GetRankValueSByOwnerIdSResponse{
		Ranks: out,
	}, nil
}

func (svc *Service) DeleteRankValue(ctx *ctx.Context, req *rank_service.RankService_DeleteRankValueRequest) (*rank_service.RankService_DeleteRankValueResponse, *errmsg.ErrMsg) {
	rank := svc.memDb.GetRank(req.RankId)
	if rank == nil {
		return nil, errmsg.NewErrRankIdNotExist()
	}

	rank.DeleteById(ctx, req.OwnerId)
	return &rank_service.RankService_DeleteRankValueResponse{}, nil
}

func (svc *Service) genRankId(typ enum.RankType) (string, *errmsg.ErrMsg) {
	// TODO RankType 合法性检查
	rankId := strconv.Itoa(int(typ)) + "@" + xid.New().String()
	return rankId, nil
}

func (svc *Service) MemRankValueChange(ctx *ctx.Context, data *rankval.MemRankValueChangeData) *errmsg.ErrMsg {
	list := make([]dao.RankValue, 0)
	for _, value := range data.AddList {
		list = append(list, dao.RankValue{
			RankId:    data.RankId,
			RankType:  value.RankType,
			OwnerId:   value.OwnerId,
			Rank:      value.Rank,
			Value1:    value.Value1,
			Value2:    value.Value2,
			Value3:    value.Value3,
			Value4:    value.Value4,
			Extra:     dao.MarshalExtra(value.Extra),
			Shard:     value.Shard,
			CreatedAt: time.Now().Unix(),
		})
	}
	for _, value := range data.UpdateList {
		list = append(list, dao.RankValue{
			RankId:   data.RankId,
			RankType: value.RankType,
			OwnerId:  value.OwnerId,
			Rank:     value.Rank,
			Value1:   value.Value1,
			Value2:   value.Value2,
			Value3:   value.Value3,
			Value4:   value.Value4,
			Extra:    dao.MarshalExtra(value.Extra),
			Shard:    value.Shard,
		})
	}
	now := time.Now().Unix()
	for _, value := range data.DeleteList {
		list = append(list, dao.RankValue{
			RankId:    data.RankId,
			RankType:  value.RankType,
			OwnerId:   value.OwnerId,
			Rank:      value.Rank,
			Value1:    value.Value1,
			Value2:    value.Value2,
			Value3:    value.Value3,
			Value4:    value.Value4,
			Extra:     dao.MarshalExtra(value.Extra),
			Shard:     value.Shard,
			DeletedAt: now,
		})
	}
	if len(list) <= 0 {
		return nil
	}
	return dao.GetDao().BatchSave(ctx, list)
}
