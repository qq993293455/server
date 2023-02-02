package gen

import (
	"sync"
	"sync/atomic"
	"time"

	"coin-server/common/orm/hashtag"
	daopb "coin-server/common/proto/dao"
	values2 "coin-server/gen-rank-server/service/gen/values"

	"go.uber.org/zap"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/gen_rank"
	"coin-server/common/proto/models"
	"coin-server/common/proto/recommend"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/gen-rank-server/service/gen/dao"
)

const (
	batchGetRoleSize = 1000
	rankSizeEachSide = 5

	oneWeekSec = 60 * 60 * 24 * 7

	maxRoutineCnt = 256
	batchDaoSize  = 200
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	sl         *values2.SkipList
	genFlag    int64
}

// 1、拿到所有用户的Id （通过子昱的服务）DONE
// 2、定时任务、即使服务挂了并且过了调度时间，服务启动后还能再次调度 DONE
// 3、简单的分组，比如每个玩家战力前后50名 DONE
// 4、延迟刷新（如果实时刷新则需要海量的事件通知，很容易造成数据不一致）
// 5、每个人的排行榜有个周期，结束后需要给玩家发奖
func NewGenRankService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		sl:         values2.NewSkipList(&values2.RankCompare{}),
		log:        log,
	}
	s.start()
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("生成排名", svc.Gen)
}

func (svc *Service) Gen(c *ctx.Context, _ *gen_rank.GenRank_GenRequest) (*gen_rank.GenRank_GenResponse, *errmsg.ErrMsg) {
	_, err := svc.gen(c)
	return &gen_rank.GenRank_GenResponse{}, err
}

func (svc *Service) start() {
	svc.svc.AfterFunc(time.Second, func(c *ctx.Context) {
		record, err := dao.GetGenRecord(c)
		if err != nil {
			svc.log.Error("Gen rank record fail", zap.Error(err))
			return
		}
		now := timer.Now()
		if now.Unix()-record.GenAt > oneWeekSec {
			cnt, er := svc.gen(c)
			if er != nil {
				svc.log.Error("Gen rank record fail", zap.Error(er))
				return
			}
			record.GenNum = cnt
			record.GenAt = now.Unix()
			dao.SaveGenRecord(c, record)
		} else {
			svc.log.Info("Rank was genned already")
		}
		// 每周一早上3点更新排行榜
		toNextSec := timer.EndOfWeek(now).Add(3*time.Hour).Unix() - now.Unix()
		svc.svc.AfterFunc(time.Duration(toNextSec)*time.Second, func(c *ctx.Context) {
			do := func(c *ctx.Context) {
				record1, err1 := dao.GetGenRecord(c)
				if err1 != nil {
					svc.log.Error("Gen rank record fail", zap.Error(err))
					return
				}
				cnt, er := svc.gen(c)
				if er != nil {
					svc.log.Error("Gen rank record fail", zap.Error(er))
					return
				}
				record1.GenNum = cnt
				record1.GenAt = timer.Now().Unix()
				dao.SaveGenRecord(c, record1)
			}
			do(c)
			svc.svc.TickFunc(oneWeekSec*time.Second, func(c *ctx.Context) bool {
				do(c)
				return true
			})
			return
		})
	})
}

func (svc *Service) gen(c *ctx.Context) (int64, *errmsg.ErrMsg) {
	if !atomic.CompareAndSwapInt64(&(svc.genFlag), 0, 1) {
		return 0, errmsg.NewErrGenIsBegin()
	}
	countResp := &recommend.Recommend_GetUserCountResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(c, 0, &recommend.Recommend_GetUserCountRequest{}, countResp); err != nil {
		return 0, err
	}
	svc.log.Info("Gen rank start", zap.Any("count", countResp.Count))
	var totalCnt int64 = 0
	for _, cnt := range countResp.Count {
		totalCnt += cnt
	}
	svc.log.Info("Get all roleId", zap.Int64("total count", totalCnt))
	roleIds := make([]string, 0, totalCnt)
	for lang, cnt := range countResp.Count {
		var start, end int64 = 0, min64(batchGetRoleSize, cnt)
		for end <= cnt && start < cnt {
			roleIdsResp := &recommend.Recommend_GetUserIdsResponse{}
			if err := svc.svc.GetNatsClient().RequestWithOut(c, 0, &recommend.Recommend_GetUserIdsRequest{
				Language: lang,
				Start:    start,
				End:      end,
			}, roleIdsResp); err != nil {
				return 0, err
			}
			roleIds = append(roleIds, roleIdsResp.RoleIds...)
			svc.log.Info("gen-ing", zap.Int64("language", lang), zap.Int64("start", start), zap.Int64("end", end), zap.Int64("count", end-start))
			start = end
			end = min64(end+batchGetRoleSize, cnt)
		}
	}
	currIdx := 0
	var wg sync.WaitGroup
	var routineCnt int64 = 0
	var eachCnt = totalCnt / hashtag.SlotNumber * 2
	var slotMap = make(map[int][]*daopb.Role)
	for currIdx < int(totalCnt) {
		role := &daopb.Role{
			RoleId: roleIds[currIdx],
		}
		slot := hashtag.Slot(role.KVKey())
		if _, exit := slotMap[slot]; !exit {
			slotMap[slot] = make([]*daopb.Role, 0, eachCnt)
		}
		slotMap[slot] = append(slotMap[slot], role)
		currIdx++
	}
	start := timer.Now()
	for slot := range slotMap {
		wg.Add(1)
		for atomic.LoadInt64(&routineCnt) >= maxRoutineCnt {
			time.Sleep(10 * time.Millisecond)
		}
		s := slot
		atomic.AddInt64(&routineCnt, 1)
		svc.svc.AfterFunc(1*time.Millisecond, func(ctx1 *ctx.Context) {
			roleInfos := slotMap[s]
			cnt := len(roleInfos)
			begin, end := 0, min(batchDaoSize, cnt)
			for end <= cnt && begin < end {
				dao.GetMultiRole(ctx1, roleInfos[begin:end])
				begin = end
				end = min(end+batchDaoSize, cnt)
			}
			for _, r := range roleInfos {
				//TODO: 判断登录时间是否过长
				svc.sl.Insert(r.RoleId, r.Power)
			}
			// svc.log.Info("Insert skipList", zap.Int("count", svc.sl.GetNodeCount()))
			wg.Done()
			atomic.AddInt64(&routineCnt, -1)
		})
	}
	wg.Wait()
	svc.log.Info("Get role info done, cost", zap.Int64("cost", time.Since(start).Milliseconds()), zap.Int("count", svc.sl.GetNodeCount()))
	node := svc.sl.FindByRank(1)
	rankMap := map[string]int64{}
	rankSize := rankSizeEachSide*2 + 1
	slCnt := svc.sl.GetNodeCount()
	for i := 0; i < rankSize; i++ {
		if i == slCnt {
			break
		}
		rankMap[node.Name()] = node.Val()
		node = node.Next()
	}
	idx := 0
	minNode := svc.sl.FindByRank(1)
	maxNode := svc.sl.FindByRank(rankSize)
	currNode := svc.sl.FindByRank(1)
	for idx < slCnt && currNode != nil {
		if idx > rankSizeEachSide+1 && idx < slCnt-rankSizeEachSide-1 {
			nxtMax := maxNode.Next()
			delete(rankMap, minNode.Name())
			rankMap[nxtMax.Name()] = nxtMax.Val()
			minNode = minNode.Next()
			maxNode = nxtMax
		}
		dao.SaveRoleRank(c, &daopb.RoleRank{
			RoleId:   currNode.Name(),
			CurrRank: rankMap,
		})
		idx++
		currNode = currNode.Next()
		if idx%1000 == 0 {
			svc.log.Info("Save rank", zap.Int("count", idx))
		}
	}
	svc.sl = values2.NewSkipList(&values2.RankCompare{})
	svc.log.Info("Gen rank end")
	atomic.StoreInt64(&(svc.genFlag), 0)
	return totalCnt, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
