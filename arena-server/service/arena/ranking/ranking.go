package ranking

import (
	arenaDao "coin-server/arena-server/service/arena/dao"
	arenaRule "coin-server/arena-server/service/arena/rule"
	"coin-server/common/ArenaRule"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/arena_service"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const ChallengeBefore = 3
const ChallengeAfter = 2
const DelTime = 86400 * 30
const RobotHead = "R:R:@:"
const pauseFight = 300

type Dirty struct {
	isDirty bool
}

type ArenaRankingRoles struct {
	Dirty
	roleId           string
	rankingId        int32
	isFirstChallenge bool
	isRobot          bool
	nickName         string
	rankingIndex     string
}

func (this_ *ArenaRankingRoles) Save(ctx *ctx.Context) {
	if !this_.isDirty {
		return
	}
	arenaDao.SaveArenaRoleInfo(ctx, this_.rankingIndex, &models.ArenaRanking_Info{
		RoleId:           this_.roleId,
		RankingId:        this_.rankingId,
		IsRobot:          this_.isRobot,
		NickName:         this_.nickName,
		IsFirstChallenge: this_.isFirstChallenge,
	})
	this_.isDirty = false
}

type ArenaRankingInfos struct {
	Dirty
	rankingIndex      string
	aType             models.ArenaType
	playerNum         int32
	robotNum          int32
	roleIdToInfo      map[string]*ArenaRankingRoles
	rankingIdtoRoleId map[int32]string
	lock              sync.RWMutex
	startTime         int64
	isClose           bool
	createTime        int64
	challengeMap      map[string]string
	challengeLockMap  map[string]int64
}

func (this_ *ArenaRankingInfos) Check(isLock bool, log *logger.Logger) {
	if isLock {
		this_.lock.RLock()
		defer this_.lock.RUnlock()
		isLock = false
	}
	// test show log
	// {
	// 	lastRankingId := this_.GetLastRankingId(isLock)
	// 	for i := int32(1); i < lastRankingId; i++ {
	// 		roleId, ok := this_.rankingIdtoRoleId[i]
	// 		if !ok {
	// 			log.Error("show rankingid error", zap.Any("rankingIndex", this_.rankingIndex), zap.Any("roleid", "nil"), zap.Any("rankingId", i))
	// 			continue
	// 		}
	// 		log.Info("show rankingid", zap.Any("rankingIndex", this_.rankingIndex), zap.Any("roleid", roleId), zap.Any("rankingId", i))
	// 	}
	// }

	lastRankingId := this_.GetLastRankingId(isLock)
	for i := int32(1); i < lastRankingId; i++ {
		_, ok := this_.rankingIdtoRoleId[i]
		if !ok {
			this_.RankingUp(isLock, i, lastRankingId, log)
			this_.isDirty = true
			break
		}
	}
	// test show log
	// {
	// 	lastRankingId = this_.GetLastRankingId(isLock)
	// 	for i := int32(1); i < lastRankingId; i++ {
	// 		roleId, ok := this_.rankingIdtoRoleId[i]
	// 		if !ok {
	// 			log.Error("show rankingid error", zap.Any("rankingIndex", this_.rankingIndex), zap.Any("roleid", "nil"), zap.Any("rankingId", i))
	// 			continue
	// 		}
	// 		log.Info("show rankingid", zap.Any("rankingIndex", this_.rankingIndex), zap.Any("roleid", roleId), zap.Any("rankingId", i))
	// 	}
	// }
}

func (this_ *ArenaRankingInfos) RankingUp(isLock bool, index int32, maxIndex int32, log *logger.Logger) {
	if isLock {
		this_.lock.RLock()
		defer this_.lock.RUnlock()
		isLock = false
	}

	addStep := int32(1)
	lastIndex := int32(0)
	for i := index + 1; i < maxIndex; i++ {
		roleId, ok := this_.rankingIdtoRoleId[i]
		if !ok {
			addStep++
			continue
		}
		roleInfo, ok := this_.roleIdToInfo[roleId]
		if !ok {
			addStep++
			continue
		}

		lastIndex = i - addStep
		this_.rankingIdtoRoleId[lastIndex] = this_.rankingIdtoRoleId[i]
		roleInfo.rankingId = lastIndex
		roleInfo.isDirty = true
		log.Info("role ranking id change", zap.Any("role", roleId), zap.Any("old ranking id", i), zap.Any("new ranking id", lastIndex))
	}
	for i := lastIndex + 1; i <= maxIndex; i++ {
		delete(this_.rankingIdtoRoleId, i)
	}
}

func (this_ *ArenaRankingInfos) LockRoleRanking(roleId string, roleRankingId int32, challengedRoleId string,
	challengedRankingId int32, arenaBattleDuration int64, log *logger.Logger) *errmsg.ErrMsg {
	this_.lock.Lock()
	defer this_.lock.Unlock()

	localChallengedRoleId, ok := this_.challengeMap[roleId]
	if ok {
		delete(this_.challengeLockMap, localChallengedRoleId)
		delete(this_.challengeMap, roleId)
	}

	challengeTime, ok := this_.challengeLockMap[challengedRoleId]
	if ok {
		if timer.Now().Unix()-challengeTime < arenaBattleDuration {
			return errmsg.NewErrArenaChallengeLock()
		}
		delete(this_.challengeLockMap, challengedRoleId)
	}

	roleInfo, ok := this_.roleIdToInfo[roleId]
	if !ok {
		log.Error("ArenaLockRoleRanking PlayerInfo not find ranking_id", zap.Any("role id", roleId), zap.Any("rankingIndex", this_.rankingIndex), zap.Any("arena type", this_.aType))
		return errmsg.NewErrArenaNotInRanking()
	}

	challengedRoleInfo, ok := this_.roleIdToInfo[challengedRoleId]
	if !ok {
		log.Error("ArenaLockRoleRanking PlayerInfo not find ranking_id", zap.Any("challengedRoleId", challengedRoleId), zap.Any("rankingIndex", this_.rankingIndex), zap.Any("arena type", this_.aType))
		return errmsg.NewErrArenaNotInRanking()
	}

	if !roleInfo.isFirstChallenge && roleInfo.rankingId != roleRankingId {
		return errmsg.NewErrArenaChallengeRankingChange()
	}

	if challengedRoleInfo.isFirstChallenge || challengedRoleInfo.rankingId != challengedRankingId {
		return errmsg.NewErrArenaChallengedRankingChange()
	}

	this_.challengeMap[roleId] = challengedRoleId
	this_.challengeLockMap[challengedRoleId] = timer.Now().Unix()
	return nil
}

func (this_ *ArenaRankingInfos) UnLockRoleRanking(roleId string, isLock bool) {
	if isLock {
		this_.lock.Lock()
		defer this_.lock.Unlock()
	}

	localChallengedRoleId, ok := this_.challengeMap[roleId]
	if ok {
		delete(this_.challengeLockMap, localChallengedRoleId)
		delete(this_.challengeMap, roleId)
	}
}

func (this_ *ArenaRankingInfos) GetLastRankingId(isLock bool) int32 {
	maxId := int32(0)
	if isLock {
		this_.lock.RLock()
		defer this_.lock.RUnlock()
	}
	for rankingId := range this_.rankingIdtoRoleId {
		if rankingId > maxId {
			maxId = rankingId
		}
	}
	return maxId + 1
}

func (this_ *ArenaRankingInfos) FindRoleId(roleId string) bool {
	this_.lock.RLock()
	defer this_.lock.RUnlock()
	_, ok := this_.roleIdToInfo[roleId]
	return ok
}

func (this_ *ArenaRankingInfos) Join(aType models.ArenaType, roleId string, lock bool, log *logger.Logger) {
	if lock {
		this_.lock.Lock()
		defer this_.lock.Unlock()
	}
	this_.playerNum++
	playerInfo := &ArenaRankingRoles{
		Dirty: Dirty{
			isDirty: true,
		},
		roleId:           roleId,
		rankingId:        0,
		isFirstChallenge: true,
		isRobot:          false,
		rankingIndex:     this_.rankingIndex,
	}
	this_.roleIdToInfo[roleId] = playerInfo
	//this_.rankingIdtoRoleId[playerInfo.rankingId] = roleId
	this_.isDirty = true
	log.Debug("player join rank", zap.Any("roleId", roleId), zap.Any("arena type", aType), zap.Any("rankingIndex", this_.rankingIndex))
}

func (this_ *ArenaRankingInfos) SaveInfos(ctx *ctx.Context, log *logger.Logger) {
	if !this_.isDirty {
		return
	}

	this_.lock.Lock()
	defer this_.lock.Unlock()

	daoInfos := &dao.ArenaRankingInfos{
		RankingIndex: this_.rankingIndex,
	}
	modelInfos := &models.ArenaRanking_Infos{
		PlayerNum: this_.playerNum,
		RobotNum:  this_.robotNum,
		StartTime: this_.startTime,
		IsClose:   this_.isClose,
	}

	for roleId, roleInfo := range this_.roleIdToInfo {
		if roleInfo.rankingId != 0 {
			rId, ok := this_.rankingIdtoRoleId[roleInfo.rankingId]
			if !ok {
				this_.rankingIdtoRoleId[roleInfo.rankingId] = roleId
			} else {
				if rId != roleId {
					isFix := false
					for tRankingId, tRoleId := range this_.rankingIdtoRoleId {
						if tRoleId == roleId {
							roleInfo.rankingId = tRankingId
							roleInfo.isDirty = true
							isFix = true
							break
						}
					}

					if !isFix {
						log.Error("player rankingid err cannot fix,  reset", zap.Any("roleId", roleId), zap.Any("RankingIndex",
							this_.rankingIndex), zap.Any("RankingIndex", this_.rankingIndex), zap.Any("OldId", roleInfo.rankingId), zap.Any("NewId", this_.playerNum+this_.robotNum))
						roleInfo.rankingId = this_.playerNum + this_.robotNum
						this_.rankingIdtoRoleId[this_.playerNum+this_.robotNum] = roleId
						roleInfo.isDirty = true
					}
				}
			}
		}
		roleInfo.Save(ctx)
	}

	daoInfos.Infos = modelInfos
	arenaDao.SaveArenaInfos(ctx, daoInfos)
	this_.isDirty = false
}

type ArenaRankingTypeData struct {
	Dirty
	aType                 models.ArenaType
	index                 uint32
	rankingInfos          map[string]*ArenaRankingInfos
	seasonSettlementTime  int64
	daySettlementTime     int64
	seasonRewardIndex     uint64
	dayRewardIndex        uint64
	lastDayRefreshTime    int64
	lastSeasonRefreshTime int64
	lockChallenge         atomic.Value
	lock                  sync.RWMutex
	dayRewardInfo         []*models.ArenaSendRewardTime
	seasonRewardInfo      []*models.ArenaSendRewardTime
}

func (this_ *ArenaRankingTypeData) TryJoinRanking(ctx *ctx.Context, aType models.ArenaType, roleId string, timeLimit int64, playerLimit int64, log *logger.Logger) (string, bool) {
	for _, infos := range this_.rankingInfos {
		rankingIndex, ok := JoinRanking(aType, roleId, infos, timeLimit, playerLimit, log)
		infos.SaveInfos(ctx, log)
		if ok {
			return rankingIndex, true
		}
	}
	return "", false
}

func (this_ *ArenaRankingTypeData) Save(ctx *ctx.Context, log *logger.Logger) {
	this_.lock.RLock()
	defer this_.lock.RUnlock()
	if !this_.isDirty {
		return
	}

	typeInfos := dao.ArenaRankingTypeInfos{
		Type: this_.aType,
		TypeInfos: &models.ArenaRanking_TypeInfos{
			Type:                 this_.aType,
			Index:                this_.index,
			SeasonSettlementTime: this_.seasonSettlementTime,
			DaySettlementTime:    this_.daySettlementTime,
			SeasonRewardIndex:    this_.seasonRewardIndex,
			DayRewardIndex:       this_.dayRewardIndex,
		},
	}

	for _, infos := range this_.rankingInfos {
		infos.SaveInfos(ctx, log)
	}

	arenaDao.SaveArenaTypeInfos(ctx, &typeInfos)
	this_.isDirty = false
}

func (this_ *ArenaRankingTypeData) SaveRewardIndex(ctx *ctx.Context, isSeansonOver bool) {
	tNow := timer.Now().Unix()
	this_.lock.Lock()
	defer this_.lock.Unlock()

	var rewardInfo *[]*models.ArenaSendRewardTime
	if isSeansonOver {
		rewardInfo = &this_.seasonRewardInfo
	} else {
		rewardInfo = &this_.dayRewardInfo
	}

	haveDel := false
	var tempRewardInfo []*models.ArenaSendRewardTime
	for _, info := range *rewardInfo {
		if info.SendRewardTime < tNow-DelTime {
			haveDel = true
			continue
		}
		tempRewardInfo = append(tempRewardInfo, info)
	}

	if haveDel {
		*rewardInfo = (*rewardInfo)[:0]
		*rewardInfo = append(*rewardInfo, tempRewardInfo...)
	}
	if isSeansonOver {
		arenaDao.SaveArenaSeasonRewardIndex(ctx, this_.aType, *rewardInfo)
	} else {

		arenaDao.SaveArenaDayRewardIndex(ctx, this_.aType, *rewardInfo)
	}
}

type ArenaRankingManager struct {
	svc             *service.Service
	log             *logger.Logger
	serverStartTime int64
	typeDatas       map[int32]*ArenaRankingTypeData
}

func NewarenaRanking(log *logger.Logger, svc *service.Service) *ArenaRankingManager {
	return &ArenaRankingManager{
		svc:       svc,
		log:       log,
		typeDatas: make(map[int32]*ArenaRankingTypeData),
	}
}

func (this_ *ArenaRankingManager) Init(serverId values.ServerId) *errmsg.ErrMsg {
	this_.serverStartTime = arenaDao.GetServerStartTime(ctx.GetContext())
	for _, aType := range models.ArenaType_value {
		if aType == int32(models.ArenaType_ArenaType_None) {
			continue
		}

		sId := ArenaRule.GetArenaServer(models.ArenaType(aType))
		if serverId != sId {
			continue
		}

		typeData := this_.LoadTypeData(models.ArenaType(aType))
		if typeData == nil {
			this_.log.Error("load type data err", zap.Any("aType", aType))
			continue
		}
		typeData.lastDayRefreshTime = typeData.daySettlementTime
		typeData.lastSeasonRefreshTime = typeData.seasonSettlementTime
		this_.typeDatas[aType] = typeData
	}
	return nil
}

func (this_ *ArenaRankingManager) Save(ctx *ctx.Context) {
	for _, typeData := range this_.typeDatas {
		typeData.Save(ctx, this_.log)
	}
}

func (this_ *ArenaRankingManager) Tick(now int64) {
	for _, aType := range models.ArenaType_value {
		if aType == int32(models.ArenaType_ArenaType_None) {
			continue
		}

		typeData, ok := this_.typeDatas[aType]
		if !ok {
			continue
		}

		if now > typeData.daySettlementTime-pauseFight {
			typeData.lockChallenge.Store(true)
			typeData.lastDayRefreshTime = typeData.daySettlementTime
		}

		if now > typeData.daySettlementTime {
			this_.SendReward(typeData, false)
		}

		if now > typeData.lastDayRefreshTime+pauseFight {
			typeData.lockChallenge.Store(false)
		}

		if now > typeData.seasonSettlementTime-pauseFight {
			typeData.lockChallenge.Store(true)
			typeData.lastSeasonRefreshTime = typeData.seasonSettlementTime
		}

		if now > typeData.seasonSettlementTime {
			this_.SendReward(typeData, true)
		}

		if now > typeData.lastSeasonRefreshTime+pauseFight {
			typeData.lockChallenge.Store(false)
		}
	}
}

func (this_ *ArenaRankingManager) SendReward(typeData *ArenaRankingTypeData, isSeasonOver bool) {
	_, ArenaDailyPrizeTime, ArenaSeasonPrizeTime, err := this_.GetNextRefreshTime(typeData.aType)
	if err != nil {
		this_.log.Error("Get Next Refresh Time Error", zap.Any("aType", typeData.aType))
		return
	}
	typeData.lock.Lock()
	if isSeasonOver {
		this_.log.Debug("send Reward Seanson", zap.Any("seasonRewardIndex", typeData.seasonRewardIndex), zap.Any("seasonSettlementTime", typeData.seasonSettlementTime))
		typeData.seasonRewardInfo = append(typeData.seasonRewardInfo, &models.ArenaSendRewardTime{
			RewardIndex:    typeData.seasonRewardIndex,
			SendRewardTime: typeData.seasonSettlementTime,
		})
		typeData.seasonSettlementTime = ArenaSeasonPrizeTime
		typeData.seasonRewardIndex++
	} else {
		this_.log.Debug("send Reward Day", zap.Any("dayRewardIndex", typeData.dayRewardIndex), zap.Any("daySettlementTime", typeData.daySettlementTime))
		typeData.dayRewardInfo = append(typeData.dayRewardInfo, &models.ArenaSendRewardTime{
			RewardIndex:    typeData.dayRewardIndex,
			SendRewardTime: typeData.daySettlementTime,
		})
		typeData.daySettlementTime = ArenaDailyPrizeTime
		typeData.dayRewardIndex++
	}
	ctxx := ctx.GetContext()
	defer ctxx.NewOrm().Do()
	typeData.isDirty = true
	typeData.lock.Unlock()
	typeData.Save(ctxx, this_.log)
	typeData.SaveRewardIndex(ctxx, isSeasonOver)
}

func (this_ *ArenaRankingManager) CreateRanking(typeData *ArenaRankingTypeData, lock bool) *ArenaRankingInfos {
	rankingIndex := strconv.FormatInt(int64(typeData.aType), 10) + ":" + strconv.FormatInt(int64(typeData.index), 10)
	infos := &ArenaRankingInfos{
		rankingIndex:      rankingIndex,
		aType:             typeData.aType,
		playerNum:         0,
		robotNum:          0,
		roleIdToInfo:      make(map[string]*ArenaRankingRoles),
		rankingIdtoRoleId: make(map[int32]string),
		startTime:         timer.Now().Unix(),
		isClose:           false,
		challengeMap:      make(map[string]string),
		challengeLockMap:  make(map[string]int64),
	}
	infos.isDirty = true

	if lock {
		typeData.lock.Lock()
		defer typeData.lock.Unlock()
	}

	typeData.index++
	typeData.rankingInfos[rankingIndex] = infos
	typeData.isDirty = true
	return infos
}

func (this_ *ArenaRankingManager) InitAllRanking(ctx *ctx.Context, typeData *ArenaRankingTypeData, rLock bool) *errmsg.ErrMsg {
	_, ArenaDailyPrizeTime, ArenaSeasonPrizeTime, err := this_.GetNextRefreshTime(typeData.aType)
	if err != nil {
		return err
	}

	typeData.daySettlementTime = ArenaDailyPrizeTime
	typeData.seasonSettlementTime = ArenaSeasonPrizeTime
	typeData.lastDayRefreshTime = ArenaDailyPrizeTime
	typeData.lastSeasonRefreshTime = ArenaSeasonPrizeTime

	if rLock {
		typeData.lock.RLock()
		defer typeData.lock.RUnlock()
	}

	for _, infos := range typeData.rankingInfos {
		if infos.robotNum == 0 {
			err := this_.InitRankingRobot(ctx, infos)
			if err != nil {
				return err
			}
			infos.isDirty = true
		}
	}
	return nil
}

func (this_ *ArenaRankingManager) InitRankingRobot(ctx *ctx.Context, infos *ArenaRankingInfos) *errmsg.ErrMsg {
	infos.lock.Lock()
	defer infos.lock.Unlock()
	configs := arenaRule.GetAllRobot(ctx)
	robotLimit, err := arenaRule.GetRobotLimit(infos.aType, this_.log)
	if err != nil {
		this_.log.Error("arenaRule.GetRobotLimit", zap.Any("infos.aType", infos.aType), zap.Any("err", err))
		return err
	}

	this_.log.Info("init ", zap.Any("infos.robotNum ", infos.robotNum), zap.Any("robotLimit", robotLimit))

	if infos.robotNum > 0 {
		return nil
	}

	for _, robot_config := range configs {
		this_.log.Info("init ", zap.Any("infos.robotNum ", infos.robotNum), zap.Any("robotLimit", robotLimit))
		if infos.robotNum >= int32(robotLimit) {
			break
		}

		name, ok := arenaRule.GetRobotName(ctx)
		if !ok {
			break
		}

		playerInfo := &ArenaRankingRoles{
			Dirty: Dirty{
				isDirty: true,
			},
			roleId:           RobotHead + strconv.FormatInt(robot_config.Id, 10),
			rankingId:        int32(robot_config.DefaultRank),
			isFirstChallenge: false,
			isRobot:          true,
			nickName:         name,
			rankingIndex:     infos.rankingIndex,
		}

		if _, ok := infos.rankingIdtoRoleId[int32(robot_config.DefaultRank)]; !ok {
			infos.rankingIdtoRoleId[int32(robot_config.DefaultRank)] = playerInfo.roleId
			infos.roleIdToInfo[playerInfo.roleId] = playerInfo
		}

		playerInfo.Save(ctx)
		infos.robotNum++
	}
	infos.isDirty = true
	return nil
}

func (this_ *ArenaRankingManager) JoinRanking(ctx *ctx.Context, aType models.ArenaType, roleId string) (string, *errmsg.ErrMsg) {
	typeData, ok := this_.typeDatas[int32(aType)]
	if !ok {
		return "", errmsg.NewErrArenaType()
	}

	rankingIndex, ok := FindAreadyJoinRanking(roleId, typeData)
	if ok {
		return rankingIndex, nil
	}

	isCreate, infos, rankingIndex, err := this_.JoinOrCreatRanking(ctx, aType, roleId, typeData)
	if err != nil {
		return "", err
	}

	if !isCreate {
		return rankingIndex, nil
	}

	infos.Join(aType, roleId, true, this_.log)
	arenaDao.SaveArenaRankingIndex(ctx, typeData.aType, infos.rankingIndex)
	typeData.Save(ctx, this_.log)
	return infos.rankingIndex, nil
}

func (this_ *ArenaRankingManager) JoinOrCreatRanking(ctx *ctx.Context, aType models.ArenaType, roleId string, typeData *ArenaRankingTypeData) (bool, *ArenaRankingInfos, string, *errmsg.ErrMsg) {
	playerLimit, err := arenaRule.GetPlayerLimit(aType, this_.log)
	if err != nil {
		return false, nil, "", err
	}

	timeLimit, err := arenaRule.GetArenaOpeningDuration(aType, this_.log)
	// test
	//timeLimit = 600
	if err != nil {
		return false, nil, "", err
	}

	typeData.lock.Lock()
	defer typeData.lock.Unlock()

	rankingIndex, ok := typeData.TryJoinRanking(ctx, aType, roleId, timeLimit, playerLimit, this_.log)
	if ok {
		return false, nil, rankingIndex, nil
	}

	infos := this_.CreateRanking(typeData, false)

	err = this_.InitAllRanking(ctx, typeData, false)
	if err != nil {
		return false, nil, "", err
	}

	return true, infos, "", nil
}

func (this_ *ArenaRankingManager) GetRankingData(aType models.ArenaType, rankingIndex string) (*ArenaRankingTypeData, *ArenaRankingInfos, *errmsg.ErrMsg) {
	typeData, ok := this_.typeDatas[int32(aType)]
	if !ok {
		return nil, nil, errmsg.NewErrArenaType()
	}
	typeData.lock.RLock()
	defer typeData.lock.RUnlock()

	infos, ok := typeData.rankingInfos[rankingIndex]
	if !ok {
		return nil, nil, errmsg.NewErrArenaRankingIndex()
	}

	return typeData, infos, nil
}

func (this_ *ArenaRankingManager) GetRanking(aType models.ArenaType, rankingIndex string, startIndex int32, endIndex int32) ([]*models.ArenaDataElement, *errmsg.ErrMsg) {
	if startIndex > endIndex || startIndex < 0 {
		return nil, errmsg.NewErrArenaRankingPos()
	}

	_, infos, err := this_.GetRankingData(aType, rankingIndex)
	if err != nil {
		return nil, err
	}

	infos.lock.RLock()
	defer infos.lock.RUnlock()
	var ret []*models.ArenaDataElement
	for i := startIndex; i < endIndex && i <= infos.playerNum+infos.robotNum; i++ {
		roleId, ok := infos.rankingIdtoRoleId[i]
		if !ok {
			continue
		}
		roleInfo, ok := infos.roleIdToInfo[roleId]
		if !ok {
			continue
		}
		ret = append(ret, &models.ArenaDataElement{
			RoleId:    roleInfo.roleId,
			RankingId: roleInfo.rankingId,
			IsRobot:   roleInfo.isRobot,
			NickName:  roleInfo.nickName,
		})
	}
	return ret, nil
}

func (this_ *ArenaRankingManager) GetSelfRanking(aType models.ArenaType, rankingIndex string, roleId string) (*arena_service.ArenaRanking_GetSelfDataResponse, *errmsg.ErrMsg) {
	typeData, infos, err := this_.GetRankingData(aType, rankingIndex)
	if err != nil {
		return nil, err
	}

	infos.lock.RLock()
	defer infos.lock.RUnlock()

	roleInfo, ok := infos.roleIdToInfo[roleId]
	if !ok {
		this_.log.Error("PlayerInfo not find ranking_id", zap.Any("role id", roleId), zap.Any("rankingIndex", rankingIndex), zap.Any("arena type", aType))
		return nil, errmsg.NewErrArenaNotInRanking()
	}

	return &arena_service.ArenaRanking_GetSelfDataResponse{
		Element: &models.ArenaDataElement{
			RoleId:           roleId,
			RankingId:        roleInfo.rankingId,
			IsRobot:          roleInfo.isRobot,
			NickName:         roleInfo.nickName,
			IsFirstChallenge: roleInfo.isFirstChallenge,
		},
		NextSettlementTime: typeData.daySettlementTime,
		SeasonOverTime:     typeData.seasonSettlementTime,
	}, nil
}

func (this_ *ArenaRankingManager) GetDailyPrizeTime(aType models.ArenaType) (int64, int64, int64, int64, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), this_.log)
		if config == nil {
			return 0, 0, 0, 0, errmsg.NewErrArenaConfig()
		}
		return int64(config.DefaultRefreshTime), int64(config.ArenaDailyPrizeTime), config.ArenaSeasonPrizeTime[0], config.ArenaSeasonPrizeTime[1], nil
	}
	return 0, 0, 0, 0, errmsg.NewErrArenaType()
}

func (this_ *ArenaRankingManager) GetNextRefreshTime(aType models.ArenaType) (int64, int64, int64, *errmsg.ErrMsg) {
	DefaultRefreshTime, ArenaDailyPrizeTime, ArenaSeasonPrizeDayTime, ArenaSeasonPrizeTime, err := this_.GetDailyPrizeTime(aType)
	if err != nil {
		return 0, 0, 0, err
	}

	nowTime := timer.BeginOfDay(timer.Now()).Unix()
	DefaultRefreshTime += nowTime
	ArenaDailyPrizeTime += nowTime
	h, m, _ := GetHMS(DefaultRefreshTime)
	DefaultRefreshTime = timer.GetNextRefreshTime(h, m)

	h, m, _ = GetHMS(ArenaDailyPrizeTime)
	ArenaDailyPrizeTime = timer.GetNextRefreshTime(h, m)

	nextSeasonDayTime := firstAndLastDate(ArenaSeasonPrizeDayTime)

	return DefaultRefreshTime, ArenaDailyPrizeTime, nextSeasonDayTime + ArenaSeasonPrizeTime, nil
}

func firstAndLastDate(day int64) int64 {
	firstDate := timer.Now().Format("2006-01") + "-" + fmt.Sprintf("%02d", day)
	middle, _ := time.ParseInLocation("2006-01-02", firstDate, timer.Now().Location())
	nextDate := middle.AddDate(0, 1, 0)
	return nextDate.UTC().Unix()
}

func (this_ *ArenaRankingManager) GetPlayerIsRobot(aType models.ArenaType, rankingIndex string, roleId string) (bool, string, *errmsg.ErrMsg) {
	_, infos, err := this_.GetRankingData(aType, rankingIndex)
	if err != nil {
		return true, "", err
	}

	infos.lock.RLock()
	defer infos.lock.RUnlock()

	playerInfo, ok := infos.roleIdToInfo[roleId]
	if !ok {
		return true, "", errmsg.NewErrArenaNotFoundPlayer()
	}
	return playerInfo.isRobot, playerInfo.nickName, nil
}

func (this_ *ArenaRankingManager) GetChanllengeRange(aType models.ArenaType, rankingIndex string, selfRoleId string) ([]*models.ArenaDataElement, *errmsg.ErrMsg) {
	typeData, infos, err := this_.GetRankingData(aType, rankingIndex)
	if err != nil {
		return nil, err
	}

	flg := typeData.lockChallenge.Load().(bool)
	if flg {
		return nil, errmsg.NewErrArenaRankingSendReward()
	}

	infos.lock.RLock()
	defer infos.lock.RUnlock()

	selfInfo, ok := infos.roleIdToInfo[selfRoleId]
	if !ok {
		return nil, errmsg.NewErrArenaNotFoundPlayer()
	}

	var rangeInfo []*models.ArenaDataElement
	totalNum := infos.GetLastRankingId(false) - 1
	if selfInfo.isFirstChallenge {
		num := int32(float32(totalNum) * 0.1)
		for i := totalNum - num; i <= totalNum; i++ {
			roleId, ok := infos.rankingIdtoRoleId[i]
			if !ok {
				continue
			}
			playerInfo, ok := infos.roleIdToInfo[roleId]
			if !ok {
				delete(infos.rankingIdtoRoleId, i)
				continue
			}

			if !playerInfo.isRobot {
				continue
			}

			if roleId == selfRoleId {
				continue
			}

			rangeInfo = append(rangeInfo, &models.ArenaDataElement{
				RoleId:    playerInfo.roleId,
				RankingId: playerInfo.rankingId,
				IsRobot:   playerInfo.isRobot,
				NickName:  playerInfo.nickName,
			})
		}
		challengeNum := ChallengeAfter + ChallengeBefore
		if len(rangeInfo) <= challengeNum {
			if len(rangeInfo) == 0 {
				selfInfo.isFirstChallenge = false
				return nil, errmsg.NewErrArenaRankingPos()
			}
			return rangeInfo, nil
		}
		for i := 1; i <= challengeNum; i++ {
			randIndex := rand.Intn(len(rangeInfo) - i)
			MoveDataToEnd(rangeInfo, randIndex, i)
		}
		return rangeInfo[len(rangeInfo)-challengeNum:], nil
	}

	minParam, maxParam, err := GetArenaMatchingParam(aType, this_.log)
	if err != nil {
		return nil, err
	}

	tempMap, ok := GetRankingRange(selfInfo.rankingId, minParam, maxParam, totalNum)
	if !ok {
		this_.log.Error("GetRankingRange", zap.Any("rankingId", selfInfo.rankingId), zap.Any("minParam", minParam), zap.Any("maxParam", maxParam), zap.Any("totalNum", totalNum))
		return nil, errmsg.NewErrArenaSystem()
	}

	for rankingId := range *tempMap {
		roleId, ok := infos.rankingIdtoRoleId[rankingId]
		if !ok {
			continue
		}

		playerInfo, ok := infos.roleIdToInfo[roleId]
		if !ok {
			delete(infos.rankingIdtoRoleId, rankingId)
			continue
		}

		rangeInfo = append(rangeInfo, &models.ArenaDataElement{
			RoleId:    playerInfo.roleId,
			RankingId: playerInfo.rankingId,
			IsRobot:   playerInfo.isRobot,
			NickName:  playerInfo.nickName,
		})
	}
	return rangeInfo, nil
}

func (this_ *ArenaRankingManager) ArenaRankingSwap(ctx *ctx.Context, aType models.ArenaType, rankingIndex string, roleId0 string, roleId1 string) (*models.RankingSwapInfo, bool, string, *errmsg.ErrMsg) {
	typeData, infos, err := this_.GetRankingData(aType, rankingIndex)
	if err != nil {
		return nil, false, "", err
	}

	role0Info, role1Info, err := GetFightRoles(infos, roleId0, roleId1, this_.log)
	if err != nil {
		return nil, false, "", err
	}

	isSetChallengeFalse := false
	if role0Info.isFirstChallenge {
		this_.ArenaRaningSetFirstChallenge(ctx, aType, rankingIndex, roleId0, true)
	}

	if role0Info.rankingId > role1Info.rankingId {
		ret := &models.RankingSwapInfo{
			Role_0OldRankingId: role0Info.rankingId,
			Role_0NewRankingId: role1Info.rankingId,
			Role_1OldRankingId: role1Info.rankingId,
			Role_1NewRankingId: role0Info.rankingId,
		}

		infos.lock.Lock()
		role0Info.rankingId, role1Info.rankingId = role1Info.rankingId, role0Info.rankingId
		infos.rankingIdtoRoleId[role0Info.rankingId] = role0Info.roleId
		infos.rankingIdtoRoleId[role1Info.rankingId] = role1Info.roleId
		infos.lock.Unlock()
		this_.log.Info("ranking id change", zap.Any("rankingindex", infos.rankingIndex), zap.Any("aRoleId", role0Info.roleId),
			zap.Any("aRankingOldId", ret.Role_0OldRankingId), zap.Any("aRankingNewId", ret.Role_0NewRankingId))
		this_.log.Info("ranking id change", zap.Any("rankingindex", infos.rankingIndex), zap.Any("aRoleId", role1Info.roleId),
			zap.Any("aRankingOldId", ret.Role_1OldRankingId), zap.Any("aRankingNewId", ret.Role_1NewRankingId))

		role0Info.isDirty, role1Info.isDirty = true, true

		role0Info.Save(ctx)
		role1Info.Save(ctx)

		arenaDao.SaveChange(ctx, aType, rankingIndex, typeData.seasonRewardIndex, typeData.dayRewardIndex, role0Info.roleId, role0Info.rankingId)
		if !role1Info.isRobot {
			arenaDao.SaveChange(ctx, aType, rankingIndex, typeData.seasonRewardIndex, typeData.dayRewardIndex, role1Info.roleId, role1Info.rankingId)
		}
		return ret, role1Info.isRobot, role1Info.nickName, nil
	}

	if isSetChallengeFalse {
		arenaDao.SaveChange(ctx, aType, rankingIndex, typeData.seasonRewardIndex, typeData.dayRewardIndex, role0Info.roleId, role0Info.rankingId)
	}

	ret := &models.RankingSwapInfo{
		Role_0OldRankingId: role0Info.rankingId,
		Role_0NewRankingId: role0Info.rankingId,
		Role_1OldRankingId: role1Info.rankingId,
		Role_1NewRankingId: role1Info.rankingId,
	}

	return ret, role1Info.isRobot, role1Info.nickName, nil
}

func (this_ *ArenaRankingManager) LoadTypeData(aType models.ArenaType) *ArenaRankingTypeData {
	typeDataPb, is_new := arenaDao.GetArenaTypeInfos(ctx.GetContext(), aType)
	// test
	//is_new = true
	if is_new {
		_, ArenaDailyPrizeTime, ArenaSeasonPrizeTime, err := this_.GetNextRefreshTime(typeDataPb.Type)
		if err != nil {
			this_.log.Error("Get Next Refresh Time Error", zap.Any("aType", typeDataPb.Type))
			panic("Get Next Refresh Time Error")
			//return nil
		}
		typeDataPb.TypeInfos.DaySettlementTime = ArenaDailyPrizeTime
		typeDataPb.TypeInfos.SeasonSettlementTime = ArenaSeasonPrizeTime
	}
	this_.typeDatas[int32(aType)] = this_.InitTypeData(aType, typeDataPb.TypeInfos)
	return this_.typeDatas[int32(aType)]
}

func (this_ *ArenaRankingManager) InitTypeData(aType models.ArenaType, typeDataPb *models.ArenaRanking_TypeInfos) *ArenaRankingTypeData {
	typeData := &ArenaRankingTypeData{
		aType:                aType,
		index:                typeDataPb.Index,
		daySettlementTime:    typeDataPb.DaySettlementTime,
		seasonSettlementTime: typeDataPb.SeasonSettlementTime,
		dayRewardIndex:       typeDataPb.DayRewardIndex,
		seasonRewardIndex:    typeDataPb.SeasonRewardIndex,
		rankingInfos:         make(map[string]*ArenaRankingInfos),
	}

	var flag bool = false
	typeData.lockChallenge.Store(flag)

	lctx := ctx.GetContext()
	defer lctx.NewOrm().Do()
	dayRewardIndex := arenaDao.LoadArenaDayRewardIndex(lctx, typeDataPb.Type)
	if dayRewardIndex != nil {
		typeData.dayRewardInfo = append(typeData.dayRewardInfo, dayRewardIndex...)
	}

	seasonRewardIndex := arenaDao.LoadArenaSeasonRewardIndex(lctx, typeDataPb.Type)
	if seasonRewardIndex != nil {
		typeData.seasonRewardInfo = append(typeData.seasonRewardInfo, seasonRewardIndex...)
	}

	rankingIndexsPb := arenaDao.LoadArenaRankingIndex(lctx, typeDataPb.Type)
	for _, rankingIndexPb := range rankingIndexsPb {
		infos, err := this_.LoadInfos(rankingIndexPb.RankingIndex, rankingIndexPb.CreateTime, typeDataPb.Type)
		if err != nil {
			this_.log.Error("load ranking index err", zap.Any("ranking index", rankingIndexPb.RankingIndex), zap.Any("errmsg", err))
			continue
		}

		infos.aType = typeData.aType
		err1 := this_.InitRankingRobot(lctx, infos)
		if err1 != nil {
			continue
		}

		infos.Check(true, this_.log)
		infos.SaveInfos(lctx, this_.log)
		typeData.rankingInfos[rankingIndexPb.RankingIndex] = infos
	}

	return typeData
}

func (this_ *ArenaRankingManager) LoadInfos(rankingIndex string, createTime int64, aType models.ArenaType) (*ArenaRankingInfos, *errmsg.ErrMsg) {
	infosPb := arenaDao.GetArenaInfos(ctx.GetContext(), rankingIndex)
	if infosPb == nil {
		return nil, errmsg.NewErrArenaRankingLoad()
	}

	infos := &ArenaRankingInfos{
		rankingIndex:      rankingIndex,
		playerNum:         infosPb.Infos.PlayerNum,
		robotNum:          infosPb.Infos.RobotNum,
		roleIdToInfo:      make(map[string]*ArenaRankingRoles),
		rankingIdtoRoleId: make(map[int32]string),
		startTime:         infosPb.Infos.StartTime,
		isClose:           infosPb.Infos.IsClose,
		createTime:        createTime,
		challengeMap:      make(map[string]string),
		challengeLockMap:  make(map[string]int64),
	}

	robotNum := int32(0)
	roleInfosPb := arenaDao.LoadArenaAllRoleInfo(ctx.GetContext(), rankingIndex)

	for _, rInfoPb := range roleInfosPb {
		roleInfoPb := rInfoPb.Info

		infos.roleIdToInfo[roleInfoPb.RoleId] = &ArenaRankingRoles{
			Dirty: Dirty{
				isDirty: false,
			},
			roleId:           roleInfoPb.RoleId,
			rankingId:        roleInfoPb.RankingId,
			isFirstChallenge: roleInfoPb.IsFirstChallenge,
			isRobot:          roleInfoPb.IsRobot,
			nickName:         roleInfoPb.NickName,
			rankingIndex:     rankingIndex,
		}

		if roleInfoPb.IsRobot {
			robotNum++
		}

		if roleInfoPb.RankingId > 0 {
			rRoleId, ok := infos.rankingIdtoRoleId[roleInfoPb.RankingId]
			if !ok {
				infos.rankingIdtoRoleId[roleInfoPb.RankingId] = roleInfoPb.RoleId
			} else {
				this_.log.Error("rankingid clash", zap.Any("ranking id", roleInfoPb.RankingId),
					zap.Any("aRoleid", rRoleId), zap.Any("aRole is Robot", infos.roleIdToInfo[rRoleId].isRobot),
					zap.Any("bRoleid", roleInfoPb.RoleId), zap.Any("bRole is Robot", roleInfoPb.IsRobot))
				if infos.roleIdToInfo[rRoleId].isRobot && !roleInfoPb.IsRobot {
					infos.rankingIdtoRoleId[roleInfoPb.RankingId] = roleInfoPb.RoleId
				} else {
					if !roleInfoPb.IsRobot {
						this_.log.Error("rankingid clash reset rankingid",
							zap.Any("Roleid", roleInfoPb.RoleId), zap.Any("rankingid", roleInfoPb.RankingId))
						infos.roleIdToInfo[roleInfoPb.RoleId].isDirty = true
						infos.roleIdToInfo[roleInfoPb.RoleId].rankingId = 0
						infos.roleIdToInfo[roleInfoPb.RoleId].isFirstChallenge = true
					}
				}
			}
		}
	}

	this_.log.Info("ranking", zap.Any("rank id", infos.rankingIndex), zap.Any("robot number", infosPb.Infos.RobotNum), zap.Any("actual robot number", robotNum))

	if robotNum < infosPb.Infos.RobotNum {
		infos.robotNum = robotNum
	}

	return infos, nil
}

func GetIndex(roleId string, aType models.ArenaType) string {
	return roleId + ":" + strconv.FormatInt(int64(aType), 10)
}

func (this_ *ArenaRankingManager) ArenaRaningSetFirstChallenge(ctx *ctx.Context, aType models.ArenaType, rankingIndex string, roleId string, flg bool) *errmsg.ErrMsg {
	typeData, infos, err := this_.GetRankingData(aType, rankingIndex)
	if err != nil {
		return err
	}

	infos.lock.Lock()
	defer infos.lock.Unlock()

	playerInfo, ok := infos.roleIdToInfo[roleId]
	if !ok {
		return errmsg.NewErrArenaNotInRanking()
	}

	if !playerInfo.isFirstChallenge {
		_, ok := infos.rankingIdtoRoleId[playerInfo.rankingId]
		if ok {
			return nil
		}
	}

	playerInfo.isFirstChallenge = false
	playerInfo.rankingId = infos.GetLastRankingId(false)
	playerInfo.isDirty = true
	playerInfo.Save(ctx)

	infos.rankingIdtoRoleId[playerInfo.rankingId] = playerInfo.roleId
	arenaDao.SaveChange(ctx, aType, rankingIndex, typeData.seasonRewardIndex, typeData.dayRewardIndex, playerInfo.roleId, playerInfo.rankingId)
	return nil
}

func (this_ *ArenaRankingManager) ArenaLockRoleRanking(ctx *ctx.Context, aType models.ArenaType, rankingIndex string, roleId string, roleRankingId int32,
	challengedRoleId string, challengedRankingId int32) *errmsg.ErrMsg {
	_, infos, err := this_.GetRankingData(aType, rankingIndex)
	if err != nil {
		return err
	}

	arenaBattleDuration, err := this_.GetBattleDuration(aType, this_.log)
	if err != nil {
		return err
	}

	return infos.LockRoleRanking(roleId, roleRankingId, challengedRoleId, challengedRankingId, arenaBattleDuration, this_.log)
}

func (this_ *ArenaRankingManager) ArenaUnLockRoleRanking(ctx *ctx.Context, aType models.ArenaType, rankingIndex string, roleId string) *errmsg.ErrMsg {
	_, infos, err := this_.GetRankingData(aType, rankingIndex)
	if err != nil {
		return err
	}

	infos.UnLockRoleRanking(roleId, true)
	return nil
}

func (this_ *ArenaRankingManager) GetBattleDuration(aType models.ArenaType, log *logger.Logger) (int64, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), log)
		if config == nil {
			return 0, errmsg.NewErrArenaConfig()
		}
		return int64(config.ArenaBattleDuration), nil
	}
	return 0, errmsg.NewErrArenaType()
}

func (this_ *ArenaRankingManager) GetRewardTime(ctx *ctx.Context) *arena_service.ArenaRanking_ArenaGetRewardTimeResponse {
	var ret *arena_service.ArenaRanking_ArenaGetRewardTimeResponse = new(arena_service.ArenaRanking_ArenaGetRewardTimeResponse)
	for aType, typeData := range this_.typeDatas {
		ret.Times = append(ret.Times, &models.ArenaRewardTime{
			Type:                 models.ArenaType(aType),
			DaySettlementTime:    typeData.daySettlementTime,
			SeasonSettlementTime: typeData.seasonSettlementTime,
		})
	}
	return ret
}

func FindAreadyJoinRanking(roleId string, typeData *ArenaRankingTypeData) (string, bool) {
	typeData.lock.RLock()
	defer typeData.lock.RUnlock()
	for _, infos := range typeData.rankingInfos {
		ok := infos.FindRoleId(roleId)
		if ok {
			return infos.rankingIndex, true
		}
	}
	return "", false
}

func JoinRanking(aType models.ArenaType, roleId string, infos *ArenaRankingInfos, timeLimit int64, playerLimit int64, log *logger.Logger) (string, bool) {
	if infos.isClose {
		return "", false
	}

	infos.lock.Lock()
	defer infos.lock.Unlock()

	if timer.Now().Unix() > infos.startTime+timeLimit {
		infos.isClose = true
		infos.isDirty = true
		return "", false
	}

	if infos.playerNum+infos.robotNum < int32(playerLimit) {
		infos.Join(aType, roleId, false, log)
		return infos.rankingIndex, true
	} else {
		infos.isClose = true
		infos.isDirty = true
	}
	return "", false
}

func GetHMS(timeUnix int64) (int, int, int) {
	local_time := time.Unix(timeUnix, 0).UTC()
	return local_time.Hour(), local_time.Minute(), local_time.Second()
}

func GetRankingRange(rankingId int32, minParam float64, maxParam float64, rankingPlayerNum int32) (*map[int32]bool, bool) {
	if rankingPlayerNum-1 < ChallengeBefore+ChallengeAfter {
		return nil, false
	}

	minRankingId := int32(float64(rankingId) * minParam)
	maxRankingId := int32(float64(rankingId) * maxParam)
	if minRankingId < 1 {
		minRankingId = 1
	}

	if maxRankingId > rankingPlayerNum {
		maxRankingId = rankingPlayerNum
	}

	minRange := rankingId - minRankingId
	lMinRange := rankingId - 1
	maxRange := maxRankingId - rankingId
	middleNum := rankingPlayerNum / 2

	beforeDiff := int32(0)
	afterDiff := int32(0)

	selectBefore := int32(ChallengeBefore)
	selectAfter := int32(ChallengeAfter)

	if minRange < ChallengeBefore {
		beforeDiff = ChallengeBefore - minRange
		selectBefore = minRange
	}

	if maxRange < ChallengeAfter {
		afterDiff = ChallengeAfter - maxRange
		selectAfter = maxRange
	}

	if beforeDiff > 0 && afterDiff == 0 {
		if lMinRange >= ChallengeBefore {
			minRankingId -= beforeDiff
			selectBefore += beforeDiff
		} else {
			maxRankingId += beforeDiff
			selectAfter += beforeDiff
			if maxRankingId > rankingPlayerNum {
				selectAfter -= maxRankingId - rankingPlayerNum
				maxRankingId = rankingPlayerNum
			}
		}
	}

	if afterDiff > 0 && beforeDiff == 0 {
		minRankingId -= afterDiff
		selectBefore += afterDiff
		if minRankingId < 1 {
			selectBefore -= 1 - minRankingId
			minRankingId = 1
		}
	}

	if beforeDiff > 0 && afterDiff > 0 {
		if rankingId < middleNum {
			minRankingId -= beforeDiff
			selectBefore += beforeDiff
			if minRankingId < 1 {
				selectBefore -= 1 - minRankingId
				minRankingId = 1
			}
			afterDiff += ChallengeBefore - (rankingId - minRankingId)
			maxRankingId += afterDiff
			selectAfter += afterDiff
			if maxRankingId > rankingPlayerNum {
				selectAfter -= maxRankingId - rankingPlayerNum
				maxRankingId = rankingPlayerNum
			}
		} else {
			maxRankingId += afterDiff
			selectAfter += afterDiff
			if maxRankingId > rankingPlayerNum {
				selectAfter -= maxRankingId - rankingPlayerNum
				maxRankingId = rankingPlayerNum
			}
			beforeDiff += ChallengeAfter - (maxRankingId - rankingId)
			minRankingId -= beforeDiff
			selectBefore += beforeDiff
			if minRankingId < 1 {
				selectBefore -= 1 - minRankingId
				minRankingId = 1
			}
		}
	}

	tempMap := make(map[int32]bool, ChallengeBefore+ChallengeAfter)

	canRand := true
	if selectBefore >= rankingId-minRankingId {
		selectBefore = rankingId - minRankingId
		canRand = false
	}

	i := int32(1)
	for i <= selectBefore {
		index := rankingId - i
		if canRand {
			index = rankingId - (int32(rand.Intn(int(rankingId-minRankingId))) + 1)
		}
		_, ok := tempMap[index]
		if ok {
			continue
		}

		if index == rankingId {
			continue
		}

		tempMap[index] = true
		i++
	}
	canRand = true
	if selectAfter > maxRankingId-rankingId {
		selectAfter = maxRankingId - rankingId
		canRand = false
	}

	i = int32(1)
	for i <= selectAfter {
		index := i + rankingId
		if canRand {
			index = rankingId + (int32(rand.Intn(int(maxRankingId-rankingId))) + 1)
		}
		_, ok := tempMap[index]
		if ok {
			continue
		}

		if index == rankingId {
			continue
		}

		tempMap[index] = true
		i++
	}

	return &tempMap, true
}

func GetArenaMatchingParam(aType models.ArenaType, log *logger.Logger) (float64, float64, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), log)
		if config == nil {
			return 0, 0, errmsg.NewErrArenaConfig()
		}
		return config.ArenaMatchingParam[0], config.ArenaMatchingParam[1], nil
	}
	return 0, 0, errmsg.NewErrArenaType()
}

func MoveDataToEnd(src []*models.ArenaDataElement, index int, offset_pos int) {
	var temp *models.ArenaDataElement = src[index]
	src[index] = src[len(src)-offset_pos]
	src[len(src)-offset_pos] = temp
}

func GetFightRoles(infos *ArenaRankingInfos, roleId0 string, roleId1 string, log *logger.Logger) (*ArenaRankingRoles, *ArenaRankingRoles, *errmsg.ErrMsg) {
	infos.lock.Lock()
	defer infos.lock.Unlock()

	role0Info, ok := infos.roleIdToInfo[roleId0]
	if !ok {
		return nil, nil, errmsg.NewErrArenaNotInRanking()
	}
	role1Info, ok := infos.roleIdToInfo[roleId1]
	if !ok {
		return nil, nil, errmsg.NewErrArenaNotInRanking()
	}

	if !role0Info.isFirstChallenge {
		rankingRole0, ok := infos.rankingIdtoRoleId[role0Info.rankingId]
		if !ok {
			log.Error("PlayerInfo not find ranking_id", zap.Any("rankingIndex", infos.rankingIndex), zap.Any("role id", roleId0), zap.Any("ranking id", role0Info.rankingId))
			delete(infos.roleIdToInfo, roleId0)
			return nil, nil, errmsg.NewErrArenaNotFoundPlayer()
		}

		if rankingRole0 != roleId0 {
			log.Error("PlayerInfo not find ranking_id", zap.Any("rankingIndex", infos.rankingIndex),
				zap.Any("RankingMap role id", rankingRole0),
				zap.Any("RoleMap role id", roleId0), zap.Any("ranking id", role0Info.rankingId))
			delete(infos.roleIdToInfo, roleId0)
			return nil, nil, errmsg.NewErrArenaNotFoundPlayer()
		}
	}

	if role1Info.isFirstChallenge {
		return nil, nil, errmsg.NewErrArenaNotFoundPlayer()
	}

	rankingRole1, ok := infos.rankingIdtoRoleId[role1Info.rankingId]
	if !ok {
		log.Error("PlayerInfo not find ranking_id", zap.Any("rankingIndex", infos.rankingIndex), zap.Any("role id", roleId1), zap.Any("ranking id", role1Info.rankingId))
		delete(infos.roleIdToInfo, roleId1)
		return nil, nil, errmsg.NewErrArenaNotFoundPlayer()
	}

	if rankingRole1 != roleId1 {
		log.Error("PlayerInfo not find ranking_id", zap.Any("rankingIndex", infos.rankingIndex),
			zap.Any("RankingMap role id", rankingRole1),
			zap.Any("RoleMap role id", roleId1), zap.Any("ranking id", role1Info.rankingId))
		delete(infos.roleIdToInfo, roleId1)
		return nil, nil, errmsg.NewErrArenaNotFoundPlayer()
	}

	return role0Info, role1Info, nil
}
