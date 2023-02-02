// Package rank 百人榜逻辑
package rank

import (
	"sort"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/rank_service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	values2 "coin-server/rank-server/service/rank/values"
)

const (
	topRankMaxCount  = 100 // 排行榜最大数量
	topRankRealCount = 120 // 实际存储数量
)

func (svc *Service) GetTopRank(ctx *ctx.Context, req *rank_service.RankService_TopRankGetRequest) (*rank_service.RankService_TopRankGetResponse, *errmsg.ErrMsg) {
	out := &dao.TopRank{Title: req.Title}
	ok, err := ctx.NewOrm().GetPB(enum.GetRedisClient(), out)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &rank_service.RankService_TopRankGetResponse{
			RoleId: nil,
		}, nil
	}
	dataList := make([]*values2.TopRankItem, 0, len(out.Data))
	for id, item := range out.Data {
		dataList = append(dataList, &values2.TopRankItem{
			RoleId:      id,
			TopRankItem: item,
		})
	}
	sort.Slice(dataList, func(i, j int) bool {
		if dataList[i].CombatValue != dataList[j].CombatValue {
			return dataList[i].CombatValue > dataList[j].CombatValue
		}
		if dataList[i].CreatedAt != dataList[j].CreatedAt {
			return dataList[i].CreatedAt < dataList[j].CreatedAt
		}
		return dataList[i].RoleId < dataList[j].RoleId
	})
	list := make([]values.RoleId, 0)
	var count int
	for _, item := range dataList {
		list = append(list, item.RoleId)
		count++
		if count >= topRankMaxCount {
			break
		}
	}
	return &rank_service.RankService_TopRankGetResponse{
		RoleId: list,
	}, nil
}

func (svc *Service) UpdateTopRank(ctx *ctx.Context, req *rank_service.RankService_TopRankUpdateRequest) (*rank_service.RankService_TopRankUpdateResponse, *errmsg.ErrMsg) {
	svc.lock.Lock()
	defer svc.lock.Unlock()
	rc := enum.GetRedisClient()
	limit := &dao.TopRankLimit{Title: req.CurrentTitle}
	orm := ctx.NewOrm()
	_, err := orm.GetPB(rc, limit)
	if err != nil {
		return nil, err
	}
	out := &dao.TopRank{Title: req.CurrentTitle}
	ok, err := orm.GetPB(rc, out)
	if err != nil {
		return nil, err
	}
	if err := svc.titleChange(ctx, req); err != nil {
		return nil, err
	}

	data := make(map[values.RoleId]*dao.TopRankItem)
	if ok && len(out.Data) > 0 {
		data = out.Data
	}
	target, ok := data[req.RoleId]
	if ok && target.CombatValue < req.CombatValue {
		data[req.RoleId].CombatValue = req.CombatValue
	} else {
		data[req.RoleId] = &dao.TopRankItem{
			CombatValue: req.CombatValue,
			CreatedAt:   time.Now().Unix(),
		}
	}
	var limitUpdate bool
	full := limit.Full
	// if req.CombatValue > limit.Min {
	// 	limit.Min = req.CombatValue
	// 	limitUpdate = true
	// }
	var min values.Integer
	if len(data) > topRankRealCount {
		out.Data, min = svc.handleTopRank(data, topRankRealCount)
	} else {
		out.Data = data
		for _, item := range data {
			if min == 0 || min > item.CombatValue {
				min = item.CombatValue
			}
		}
	}
	if limit.Min != min {
		limit.Min = min
		limitUpdate = true
	}

	limit.Full = len(data) >= topRankRealCount
	if limitUpdate || limit.Full != full {
		orm.SetPB(rc, limit)
	}
	orm.SetPB(rc, out)

	return &rank_service.RankService_TopRankUpdateResponse{}, nil
}

func (svc *Service) handleTopRank(data map[values.RoleId]*dao.TopRankItem, max int) (map[values.RoleId]*dao.TopRankItem, values.Integer) {
	list := make([]*values2.TopRankItem, 0, len(data))
	for id, item := range data {
		list = append(list, &values2.TopRankItem{
			RoleId: id,
			TopRankItem: &dao.TopRankItem{
				CombatValue: item.CombatValue,
				CreatedAt:   item.CreatedAt,
			},
		})
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].CombatValue != list[j].CombatValue {
			return list[i].CombatValue > list[j].CombatValue
		} else {
			return list[i].CreatedAt < list[j].CreatedAt
		}
	})
	list = list[:max]
	ret := make(map[values.RoleId]*dao.TopRankItem)
	for _, item := range list {
		ret[item.RoleId] = item.TopRankItem
	}
	return ret, list[len(list)-1].CombatValue
}

func (svc *Service) titleChange(ctx *ctx.Context, req *rank_service.RankService_TopRankUpdateRequest) *errmsg.ErrMsg {
	if req.LastTitle == req.CurrentTitle {
		return nil
	}
	// 头衔变化，将玩家从旧的头衔排行榜里移除
	rc := enum.GetRedisClient()
	orm := ctx.NewOrm()
	out := &dao.TopRank{Title: req.LastTitle}
	ok, err := orm.GetPB(rc, out)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if len(out.Data) <= 0 {
		return nil
	}
	if _, ok := out.Data[req.RoleId]; !ok {
		return nil
	}
	delete(out.Data, req.RoleId)
	orm.SetPB(rc, out)
	return nil
}
