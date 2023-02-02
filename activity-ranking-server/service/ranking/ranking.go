package activity_ranking

import (
	"sort"
	"strconv"
	"sync"

	//aServiceRankPb "coin-server/common/proto/activity_ranking_service"
	aDao "coin-server/activity-ranking-server/service/ranking/dao"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/activity_ranking_service"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"

	"go.uber.org/zap"
)

type Dirty struct {
	isDirty bool
}

type RankingData struct {
	Dirty
	models.ActivityRanking_Data
	version int64
}

type RankingInfo struct {
	Dirty
	models.ActivityRanking_Info

	roleLock     sync.RWMutex
	roleIdToInfo map[string]*RankingData
	rankingIndex string // 字段头 需要外围指定

	rankingIdtoRoleId RankingList
	isEnd             bool
	lastRefresh       int64
}

func (this_ *RankingInfo) Init(ctx *ctx.Context) {
	this_.rankingIdtoRoleId.Init()
	datas := aDao.GetActivityRankingData(ctx, this_.rankingIndex, this_.Version)

	for _, data := range datas {
		sData := &RankingData{
			Dirty:                Dirty{isDirty: false},
			ActivityRanking_Data: *data.Data,
		}
		this_.roleIdToInfo[data.RoleId] = sData
	}

	this_.Sort(true, timer.Now().Unix())
}

func (this_ *RankingInfo) Save(ctx *ctx.Context, roleLock *LocalMutex) {
	if roleLock == nil {
		roleLock = this_.GetRoleLock()
		defer roleLock.UnLock()
	}
	roleLock.RLock()

	for _, data := range this_.roleIdToInfo {
		if data.isDirty {
			aDao.SaveActivityRankingData(ctx, this_.rankingIndex, this_.Version, &data.ActivityRanking_Data)
			data.isDirty = false
		}
	}
	this_.isDirty = false
}

func (this_ *RankingInfo) Delete(ctx *ctx.Context) {
	ctx.Info("delet Ranking", zap.String("rankingIndex", this_.rankingIndex))
	roleLock := this_.GetRoleLock()
	roleLock.WLock()
	defer roleLock.UnLock()
	for key, data := range this_.roleIdToInfo {
		aDao.DelActivityRankingData(ctx, this_.rankingIndex, this_.Version, data.RoleId)
		delete(this_.roleIdToInfo, key)
	}
}

func (this_ *RankingInfo) Join(ctx *ctx.Context, roleId string, score int64) {
	roleLock := this_.GetRoleLock()
	roleLock.WLock()
	defer roleLock.UnLock()

	_, ok := this_.roleIdToInfo[roleId]
	if ok {
		return
	}

	data := &RankingData{
		Dirty: Dirty{isDirty: true},
		ActivityRanking_Data: models.ActivityRanking_Data{
			RoleId:     roleId,
			Score:      score,
			UpdateTime: timer.Now().Unix(),
		},
	}

	this_.roleIdToInfo[roleId] = data
	this_.SetDirty()
	this_.Save(ctx, roleLock)
}

func (this_ *RankingInfo) Update(ctx *ctx.Context, roleId string, score int64) bool {
	roleLock := this_.GetRoleLock()
	roleLock.RLock()
	defer roleLock.UnLock()

	data, ok := this_.roleIdToInfo[roleId]
	if !ok {
		data = &RankingData{
			Dirty: Dirty{isDirty: true},
			ActivityRanking_Data: models.ActivityRanking_Data{
				RoleId:     roleId,
				Score:      score,
				UpdateTime: timer.Now().Unix(),
			},
		}
		this_.roleIdToInfo[roleId] = data
		this_.SetDirty()
	} else {
		rankingId := this_.GetRankingId(roleId)
		data.RankingId = rankingId
		data.NewScore = score
	}

	this_.SetDataDirty(data)
	this_.Save(ctx, roleLock)

	return true
}

func (this_ *RankingInfo) GetRankingId(roleId string) int64 {
	data := this_.rankingIdtoRoleId.GetReadData()
	defer data.UnLock()
	dataMap := data.GetReadMap()
	rankingId, ok := dataMap.RoleIdTorankingId[roleId]
	if !ok {
		return -1
	}
	return rankingId
}

func (this_ *RankingInfo) GetList(startIndex int64, cnt int64) ([]*models.ActivityRanking_RankInfo, int64) {
	data := this_.rankingIdtoRoleId.GetReadData()
	defer data.UnLock()
	dataMap := data.GetReadMap()
	nextRefreshTime := this_.RefreshTime + this_.lastRefresh

	maxLen := int64(len(dataMap.RankingIdToRoleId))
	if startIndex > maxLen {
		return nil, nextRefreshTime
	}

	ret := make([]*models.ActivityRanking_RankInfo, 0, cnt)

	endIndex := startIndex + cnt
	for i := startIndex; i < endIndex && i < maxLen; i++ {
		roleId, ok := dataMap.RankingIdToRoleId[i]
		if !ok {
			continue
		}
		rd := this_.roleIdToInfo[roleId]
		ret = append(ret, &models.ActivityRanking_RankInfo{
			RoleId:    roleId,
			RankingId: int32(i),
			Score:     rd.Score,
		})
	}
	return ret, nextRefreshTime
}

func (this_ *RankingInfo) UpdateEndInfo(ctx *ctx.Context) {
	roleLock := this_.GetRoleLock()
	roleLock.RLock()
	defer roleLock.UnLock()

	data := this_.rankingIdtoRoleId.GetReadData()
	defer data.UnLock()
	dataMap := data.GetReadMap()

	for _, data := range this_.roleIdToInfo {
		rankingId, ok := dataMap.RoleIdTorankingId[data.RoleId]
		if !ok {
			rankingId = -1
		}
		data.RankingId = rankingId
		data.IsOver = true
		data.isDirty = true
	}
	this_.isDirty = true
	this_.Save(ctx, roleLock)
}

func (this_ *RankingInfo) Sort(isForceRefresh bool, timeNow int64) {
	if !isForceRefresh {
		if this_.RefreshTime+this_.lastRefresh > timeNow {
			return
		}
	}
	this_.lastRefresh = timeNow

	if !this_.rankingIdtoRoleId.isDirty {
		return
	}

	roleLock := this_.GetRoleLock()
	roleLock.RLock()
	defer roleLock.UnLock()

	var datas []*RankingData
	for _, data := range this_.roleIdToInfo {
		if data.NewScore > data.Score {
			data.Score = data.NewScore
		}
		datas = append(datas, data)
	}
	roleLock.UnLock()

	sort.Slice(datas, func(i, j int) bool {
		if datas[i].Score > datas[j].Score {
			return true
		}
		if datas[i].Score == datas[j].Score {
			if datas[i].UpdateTime < datas[j].UpdateTime {
				return true
			}
			if datas[i].UpdateTime == datas[j].UpdateTime {
				return datas[i].RoleId > datas[j].RoleId
			}
		}
		return false
	})

	WriteData := this_.rankingIdtoRoleId.GetWriteData()
	defer WriteData.UnLock()
	rankingDataMap := WriteData.GetClearWriteMap()
	for index, data := range datas {
		if data.RankingId != int64(index) {
			data.RankingId = int64(index)
			data.isDirty = true
		}
		rankingDataMap.RankingIdToRoleId[int64(index)] = data.RoleId
		rankingDataMap.RoleIdTorankingId[data.RoleId] = int64(index)
	}
	ReadData := this_.rankingIdtoRoleId.GetReadData()
	defer ReadData.UnLock()
	Swap(ReadData, WriteData)
	this_.rankingIdtoRoleId.isDirty = false

	c := ctx.GetContext()
	this_.Save(c, nil)
	err := c.GetOrmForMiddleWare().Do()
	if err != nil {
		c.Error("call Save error", zap.Error(err))
	}
}

func (this_ *RankingInfo) GetRoleLock() *LocalMutex {
	return GetLocalLock(&this_.roleLock)
}

func (this_ *RankingInfo) SetDirty() {
	this_.isDirty = true
	this_.SetShowDirty()
}

func (this_ *RankingInfo) SetShowDirty() {
	this_.rankingIdtoRoleId.isDirty = true
}

func (this_ *RankingInfo) SetDataDirty(data *RankingData) {
	data.isDirty = true
	data.UpdateTime = timer.Now().Unix()
	this_.SetShowDirty()
}

type RankingManager struct {
	Dirty
	svc          *service.Service
	log          *logger.Logger
	lock         sync.RWMutex
	ServerId     values.ServerId
	rankingInfos map[string]*RankingInfo
}

func NewActivityRanking(log *logger.Logger, svc *service.Service) *RankingManager {
	return &RankingManager{
		svc:          svc,
		log:          log,
		rankingInfos: make(map[string]*RankingInfo),
	}
}

func (this_ *RankingManager) Init(ctx *ctx.Context, serverId values.ServerId) *errmsg.ErrMsg {
	timeNow := timer.Now().Unix()
	this_.ServerId = serverId
	managerData, _ := aDao.GetActivityManagerInfo(ctx, serverId)
	for rankingIndex, info := range managerData.Infos {
		if info.EndTime != -1 && info.EndTime < timeNow {
			ctx.Info("ranking is time out delete", zap.String("ranking index", rankingIndex))
			continue
		}

		sInfo := &RankingInfo{
			Dirty: Dirty{
				isDirty: false,
			},
			ActivityRanking_Info: *info,
			roleIdToInfo:         make(map[string]*RankingData),
			rankingIndex:         rankingIndex,
		}
		sInfo.Init(ctx)
		this_.rankingInfos[rankingIndex] = sInfo
	}
	return nil
}

func (this_ *RankingManager) Tick(timeNow int64) *errmsg.ErrMsg {
	lock := this_.GetLock()
	lock.RLock()
	defer lock.UnLock()

	var ctxx *ctx.Context
	for _, info := range this_.rankingInfos {
		isForceRefresh := false
		if info.isEnd {
			continue
		}
		if info.EndTime != -1 && info.EndTime < timeNow {
			info.isEnd = true
			isForceRefresh = true
		}
		info.Sort(isForceRefresh, timeNow)
		if info.isEnd {
			if ctxx == nil {
				defer ctxx.NewOrm().Do()
				ctxx = ctx.GetContext()
			}
			info.UpdateEndInfo(ctxx)
		}
	}

	return nil
}

func (this_ *RankingManager) Save(ctx *ctx.Context, lock *LocalMutex) {
	if !this_.isDirty {
		return
	}

	if lock == nil {
		lock = this_.GetLock()
		defer lock.UnLock()
	}
	lock.RLock()

	data := &dao.ActivityManagerInfo{
		ServerId: strconv.FormatInt(this_.ServerId, 10),
		Infos:    make(map[string]*models.ActivityRanking_Info),
	}
	for _, info := range this_.rankingInfos {
		data.Infos[info.rankingIndex] = &info.ActivityRanking_Info
		info.Save(ctx, nil)
	}
	aDao.SaveActivityManagerInfo(ctx, data)
	this_.isDirty = false
}

func (this_ *RankingManager) Create(ctx *ctx.Context, rankingIndex string, createInfo *models.ActivityRanking_Info, lock *LocalMutex) *RankingInfo {
	if lock == nil {
		lock = this_.GetLock()
		defer lock.UnLock()
	}
	lock.WLock()
	rankingInfo, ok := this_.rankingInfos[rankingIndex]
	if ok {
		return rankingInfo
	}
	sInfo := &RankingInfo{
		Dirty: Dirty{
			isDirty: false,
		},
		ActivityRanking_Info: *createInfo,
		roleIdToInfo:         make(map[string]*RankingData),
		rankingIndex:         rankingIndex,
	}
	sInfo.Init(ctx)
	this_.rankingInfos[rankingIndex] = sInfo
	this_.isDirty = true
	return sInfo
}

func (this_ *RankingManager) Join(ctx *ctx.Context, joinData *activity_ranking_service.ActivityRanking_JoinRankingRequest) (string, *errmsg.ErrMsg) {
	lock := this_.GetLock()
	lock.RLock()
	defer lock.UnLock()

	rankingIndex := joinData.RankingIndex + ":" + strconv.FormatInt(joinData.Version, 10)

	rankingInfo, ok := this_.rankingInfos[rankingIndex]
	if !ok {
		if joinData.CreateInfo.RefreshTime == 0 {
			return "", errmsg.NewErrActivityRankingCreateConfig()
		}
		rankingInfo = this_.Create(ctx, rankingIndex, joinData.CreateInfo, lock)
		if joinData.Score > 0 {
			rankingInfo.Join(ctx, joinData.RoleId, joinData.Score)
		}
		this_.Save(ctx, lock)
		return rankingIndex, nil
	}

	if rankingInfo.isEnd {
		return "", errmsg.NewErrActivityRankingIsOver()
	}

	if joinData.Score > 0 {
		rankingInfo.Join(ctx, joinData.RoleId, joinData.Score)
	}
	return rankingIndex, nil
}

func (this_ *RankingManager) Update(ctx *ctx.Context, updateData *activity_ranking_service.ActivityRanking_UpdateRankingDataRequest) *errmsg.ErrMsg {
	lock := this_.GetLock()
	lock.RLock()
	defer lock.UnLock()

	rankingInfo, ok := this_.rankingInfos[updateData.RankingIndex]
	if !ok {
		return errmsg.NewErrActivityRankingNotRanking()
	}

	if rankingInfo.isEnd {
		return errmsg.NewErrActivityRankingIsOver()
	}

	rankingInfo.Update(ctx, updateData.RoleId, updateData.Score)
	return nil
}

func (this_ *RankingManager) GetSelfRankingInfo(ctx *ctx.Context, req *activity_ranking_service.ActivityRanking_GetSelfRankingRequest) (*activity_ranking_service.ActivityRanking_GetSelfRankingResponse, *errmsg.ErrMsg) {
	lock := this_.GetLock()
	lock.RLock()
	defer lock.UnLock()

	rankingInfo, ok := this_.rankingInfos[req.RankingIndex]
	if !ok {
		return nil, errmsg.NewErrActivityRankingNotRanking()
	}

	rankingId := rankingInfo.GetRankingId(req.RoleId)

	return &activity_ranking_service.ActivityRanking_GetSelfRankingResponse{
		RankingId: rankingId,
	}, nil
}

func (this_ *RankingManager) GetRankingList(ctx *ctx.Context, req *activity_ranking_service.ActivityRanking_GetRankingListRequest) (*activity_ranking_service.ActivityRanking_GetRankingListResponse, *errmsg.ErrMsg) {
	lock := this_.GetLock()
	lock.RLock()
	defer lock.UnLock()

	rankingInfo, ok := this_.rankingInfos[req.RankingIndex]
	if !ok {
		return nil, errmsg.NewErrActivityRankingNotRanking()
	}

	rankingList, nextRefreshTime := rankingInfo.GetList(req.StartIndex, req.Count)

	rankingId := rankingInfo.GetRankingId(req.RoleId)

	return &activity_ranking_service.ActivityRanking_GetRankingListResponse{
		RankingList:     rankingList,
		SelfRankingId:   rankingId,
		NextRefreshTime: nextRefreshTime,
	}, nil
}

func (this_ *RankingManager) GetLock() *LocalMutex {
	return GetLocalLock(&this_.lock)
}
