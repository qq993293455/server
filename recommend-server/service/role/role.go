package role

import (
	"math/rand"
	"sync"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/proto/recommend"
	"coin-server/common/routine_limit_service"
	"coin-server/common/values"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *routine_limit_service.RoutineLimitService
	log        *logger.Logger
	roleMap    map[int64]*LangRole
	roleCnt    values.Integer
	lock       *sync.RWMutex
}

func NewRoleService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *routine_limit_service.RoutineLimitService,
	roleMap map[int64]*LangRole,
	roleCnt values.Integer,
	log *logger.Logger,
	lock *sync.RWMutex,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		roleMap:    roleMap,
		roleCnt:    roleCnt,
		lock:       lock,
		log:        log,
	}
	return s
}

func (svc *Service) Router() {
	//svc.svc.RegisterEvent("新增角色", svc.AddRole)
	svc.svc.RegisterFunc("获取推荐好友", svc.Recommend)
	svc.svc.RegisterFunc("获取各语言玩家人数", svc.GetUserCount)
	svc.svc.RegisterFunc("获取玩家id", svc.GetUserIds)
}

// 改为每日刷新
//func (svc *Service) AddRole(ctx *ctx.Context, req *recommend.Recommend_UserEnterEvent) {
//	roles, ok := svc.roleMap[req.Role.Language]
//	if !ok {
//		svc.lock.Lock()
//		svc.roleMap[req.Role.Language] = &LangRole{
//			RoleList: make([]Role, 0),
//			Lock:     &sync.RWMutex{},
//		}
//		svc.lock.Unlock()
//	}
//	roles.Lock.Lock()
//	roles.RoleList = append(roles.RoleList, NewRole((*dao.Role)(req.Role)))
//	svc.roleCnt++
//	roles.Lock.Unlock()
//	return
//}

func (svc *Service) Recommend(ctx *ctx.Context, req *recommend.Recommend_RecommendRequest) (*recommend.Recommend_RecommendResponse, *errmsg.ErrMsg) {
	var list []string
	num := 10
	if req.Language == 0 {
		return nil, nil
	}

	// 总人数小于等于推荐的人数，推荐全部人
	svc.lock.RLock()
	if int(svc.roleCnt) <= num {
		if int(svc.roleCnt) == 0 {
			return nil, nil
		}
		for _, v := range svc.roleMap {
			for idx := range v.RoleList {
				if ctx.RoleId == v.RoleList[idx].RoleId {
					continue
				}
				list = append(list, v.RoleList[idx].RoleId)
			}
		}
		svc.lock.RUnlock()
	} else {
		svc.lock.RUnlock()
		//rand.Seed(time.Now().UnixNano())
		list = svc.doRecommendN(ctx, num, req.Language)
		// 本语言人数不足推荐人数, 推荐其他语言
		if len(list) < num {
			other := svc.doRecommendOther(ctx, num-len(list), req.Language)
			list = append(list, other...)
		}
	}

	return &recommend.Recommend_RecommendResponse{
		Id: list,
	}, nil
}

func (svc *Service) GetUserCount(ctx *ctx.Context, _ *recommend.Recommend_GetUserCountRequest) (*recommend.Recommend_GetUserCountResponse, *errmsg.ErrMsg) {
	svc.lock.RLock()
	defer svc.lock.RUnlock()
	m := map[int64]values.Integer{}
	for k, v := range svc.roleMap {
		m[k] = values.Integer(len(v.RoleList))
	}
	return &recommend.Recommend_GetUserCountResponse{Count: m}, nil
}

func (svc *Service) GetUserIds(ctx *ctx.Context, req *recommend.Recommend_GetUserIdsRequest) (*recommend.Recommend_GetUserIdsResponse, *errmsg.ErrMsg) {
	svc.lock.RLock()
	defer svc.lock.RUnlock()
	roles, ok := svc.roleMap[req.Language]
	if !ok {
		return nil, nil
	}
	if req.Start < 0 || req.End < 0 || req.Start > req.End {
		return nil, nil
	}
	if req.End-req.Start > 1000 {
		return nil, errmsg.NewErrRecommendNumTooLarge()
	}
	if len(roles.RoleList) < int(req.End) {
		req.End = int64(len(roles.RoleList))
	}

	ids := make([]string, 0)
	for i := req.Start; i < req.End; i++ {
		ids = append(ids, roles.RoleList[i].RoleId)
	}

	return &recommend.Recommend_GetUserIdsResponse{RoleIds: ids}, nil
}

func (svc *Service) doRecommendN(ctx *ctx.Context, num int, lang int64) []string {
	roles, ok := svc.roleMap[lang]
	if !ok {
		return nil
	}
	if len(roles.RoleList) == 0 {
		return nil
	}
	var list []string

	roles.Lock.RLock()
	defer roles.Lock.RUnlock()
	if len(roles.RoleList) <= num {
		list = make([]string, 0)
		for i := 0; i < len(roles.RoleList); i++ {
			if ctx.RoleId != roles.RoleList[i].RoleId {
				list = append(list, roles.RoleList[i].RoleId)
			}
		}
		return list
	}

	list = make([]string, num)
	randMap := make(map[int]struct{})
	// TODO 等策划需求
	for i := 0; i < num; i++ {
		r := rand.Intn(len(roles.RoleList))
		// 随到重复，重新随
		if _, ok := randMap[r]; ok {
			i--
			continue
		}
		// 随到自己，重新随
		if ctx.RoleId == roles.RoleList[i].RoleId {
			i--
			continue
		}
		list[i] = roles.RoleList[i].RoleId
		randMap[r] = struct{}{}
	}
	return list
}

// 随机推荐其他语种（公平）
func (svc *Service) doRecommendOther(_ *ctx.Context, num int, lang int64) []string {
	list := make([]string, 0)
	svc.lock.RLock()
	defer svc.lock.RUnlock()
	for i := 0; i < num; i++ {
		r := rand.Intn(len(svc.roleMap))
		n := 0
		for k, v := range svc.roleMap {
			if n == r {
				// 本语言，重新随
				if k == lang {
					i--
					break
				}
				rr := rand.Intn(len(v.RoleList))
				list = append(list, v.RoleList[rr].RoleId)
				break
			}
			n++
		}
	}
	return list
}

func (svc *Service) doRecommend(ctx *ctx.Context, lang int64) string {
	roles, ok := svc.roleMap[lang]
	if !ok {
		return ""
	}
	// 保证本语言不只自己一人
	if len(roles.RoleList) < 2 {
		return ""
	}

	roles.Lock.RLock()
	defer roles.Lock.RUnlock()

	// TODO 等策划需求
	r := rand.Intn(len(roles.RoleList))
	// 保证能取到一个不是自己的id
	if ctx.RoleId == roles.RoleList[r].RoleId {
		if r == len(roles.RoleList)-1 {
			return roles.RoleList[r-1].RoleId
		} else {
			return roles.RoleList[r+1].RoleId
		}
	}
	return roles.RoleList[r].RoleId
}

// TODO 临时，为了压测
func randName() string {
	str := "abcdefghijklmnopqrstuvwxyz0123456789"
	n := rand.Intn(len(str))
	return str[n : n+1]
}
