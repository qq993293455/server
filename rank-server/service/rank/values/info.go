package values

import (
	"fmt"
	"sync"

	"coin-server/common/ctx"
	"coin-server/common/proto/models"
	"coin-server/common/skiplist"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/enum"
)

type RankAgg interface {
	// 排行榜ID
	GetRankId() values.RankId
	// 排行榜类型
	GetRankType() enum.RankType
	// 获取排行榜总个数
	GetTotalNum() values.Integer
	// 获取指定Id的详细信息
	GetValueById(ownerId values.GuildId) *models.RankValue
	// 批量获取指定id的排行榜积分
	GetScoreValue1ByIds(ownerIds []values.GuildId) map[values.GuildId]values.Integer
	GetScoreValue2ByIds(ownerIds []values.GuildId) map[values.GuildId]models.RankAndScore
	// 获取指定排名的详细信息
	GetValueByRank(rank values.Integer) *models.RankValue
	// 按排名范围获取排行榜上的详细信息,范围是闭区间，比如 start 为1， end为2，返回的是第一名和第二名
	GetValueByRange(start, end values.Integer) []*models.RankValue
	InitValue(value *models.RankValue)
	// 更新排行榜上的数据
	UpdateValue(ctx *ctx.Context, value *models.RankValue)
	// 删除排行榜上的数据
	DeleteById(ctx *ctx.Context, ownerId values.GuildId)
	//
	ClearAll()
}

type RankInfo struct {
	RankId   values.RankId
	rankType enum.RankType
	sl       *skiplist.SkipList
	lock     sync.RWMutex
}

func NewRankInfo(rankId values.RankId, rankType enum.RankType, valuesArray []*models.RankValue) RankAgg {
	f, ok := AllRankFuncs[rankType]
	if !ok {
		utils.Must(fmt.Errorf("invalid sort type %d", rankType))
	}
	compare := &RankCompare{
		F: f,
	}
	res := &RankInfo{
		RankId:   rankId,
		rankType: rankType,
		sl:       skiplist.NewSkipList(compare),
	}
	for i := range valuesArray {
		tmp := valuesArray[i]
		res.sl.Insert(tmp.OwnerId, tmp)
	}
	return res
}

func (r *RankInfo) GetRankId() values.RankId {
	return r.RankId
}

func (r *RankInfo) GetRankType() enum.RankType {
	return r.rankType
}

func (r *RankInfo) GetTotalNum() values.Integer {
	r.lock.RLock()
	defer r.lock.RUnlock()
	length := values.Integer(r.sl.GetNodeCount())
	return length
}

func (r *RankInfo) GetAllNum() values.Integer {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return values.Integer(r.sl.GetNodeCount())
}

func (r *RankInfo) GetValueById(ownerId values.GuildId) *models.RankValue {
	r.lock.RLock()
	defer r.lock.RUnlock()
	v, ok := r.sl.Get(ownerId)
	if !ok {
		return nil
	}
	rank, exist := r.sl.GetRank(ownerId)
	if !exist {
		utils.Must(fmt.Errorf("can not find value in skip list,  %+v", v))
	}

	tmp := v.(*models.RankValue)
	res := &models.RankValue{
		RankId:  tmp.RankId,
		OwnerId: tmp.OwnerId,
		Rank:    values.Integer(rank),
		Value1:  tmp.Value1,
		Value2:  tmp.Value2,
		Value3:  tmp.Value3,
		Value4:  tmp.Value4,
		Extra:   make(map[string]string),
	}
	for k, v := range tmp.Extra {
		res.Extra[k] = v
	}
	// deepcopier.DeepCopy(res, v)
	// res.Rank = identify.Integer(rank)
	return res
}

func (r *RankInfo) GetScoreValue1ByIds(ownerIds []values.GuildId) map[values.GuildId]values.Integer {
	r.lock.RLock()
	defer r.lock.RUnlock()
	out := make(map[values.GuildId]values.Integer, len(ownerIds))
	for _, ownerId := range ownerIds {
		v, ok := r.sl.Get(ownerId)
		if !ok {
			out[ownerId] = 0
			continue
		}
		out[ownerId] = v.(*models.RankValue).Value1
	}
	return out
}

func (r *RankInfo) GetScoreValue2ByIds(ownerIds []values.GuildId) map[values.GuildId]models.RankAndScore {
	r.lock.RLock()
	defer r.lock.RUnlock()
	out := make(map[values.GuildId]models.RankAndScore, len(ownerIds))
	for _, ownerId := range ownerIds {
		v, ok := r.sl.Get(ownerId)
		if !ok {
			out[ownerId] = models.RankAndScore{}
			continue
		}
		value := v.(*models.RankValue)
		ras := models.RankAndScore{
			OwnerId: ownerId,
			Score:   value.Value1,
		}
		rank, ok := r.sl.GetRank(ownerId)
		if ok {
			ras.Rank = values.Integer(rank)
		}
		out[ownerId] = ras
	}
	return out
}

func (r *RankInfo) GetValueByRank(rank values.Integer) *models.RankValue {
	r.lock.RLock()
	defer r.lock.RUnlock()
	element := r.sl.FindByRank(int(rank))
	if element != nil && element.GetVal() != nil {
		v := element.GetVal().(*models.RankValue)
		res := &models.RankValue{}
		utils.DeepCopy(res, v)
		res.Rank = rank
		return res
	}
	return nil
}

func (r *RankInfo) GetValueByRange(start, end values.Integer) []*models.RankValue {
	if start < 0 {
		return nil
	}
	if end < start {
		start, end = end, start
	}
	realStart, realEnd := start, end
	needCount := realEnd - realStart + 1
	rvs := make([]*models.RankValue, 0, needCount)
	r.lock.RLock()
	defer r.lock.RUnlock()
	for ele := r.sl.FindByRank(int(realStart)); ele != nil && needCount > 0; ele = ele.Next() {
		v := ele.GetVal().(*models.RankValue)
		res := &models.RankValue{}
		utils.DeepCopy(res, v)
		res.Rank = realEnd - needCount + 1
		needCount--
		rvs = append(rvs, res)
	}
	return rvs
}

func (r *RankInfo) DeleteById(ctx *ctx.Context, ownerId values.GuildId) {
	r.lock.Lock()
	defer r.lock.Unlock()
	v, ok := r.sl.Get(ownerId)
	if !ok {
		return
	}
	r.sl.Delete(ownerId)
	msg := &MemRankValueChangeData{
		RankId:     r.RankId,
		DeleteList: []*models.RankValue{v.(*models.RankValue)},
	}
	ctx.PublishEventLocal(msg)
}

func (r *RankInfo) InitValue(val *models.RankValue) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.sl.Insert(val.OwnerId, val)
}

func (r *RankInfo) UpdateValue(ctx *ctx.Context, val *models.RankValue) {
	r.lock.Lock()
	defer r.lock.Unlock()
	msg := &MemRankValueChangeData{
		RankId: r.RankId,
	}
	if _, ok := r.sl.Get(val.OwnerId); !ok {
		msg.AddList = append(msg.AddList, val)
	} else {
		msg.UpdateList = append(msg.UpdateList, val)
	}
	r.sl.Insert(val.OwnerId, val)
	ctx.PublishEventLocal(msg)
}

func (r *RankInfo) ClearAll() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.sl = nil
}
