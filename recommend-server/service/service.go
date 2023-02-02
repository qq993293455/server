package service

import (
	"context"
	"sync"
	"time"

	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/orm"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/routine_limit_service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/recommend-server/service/role"

	"go.uber.org/zap"
)

type Service struct {
	svc        *routine_limit_service.RoutineLimitService
	log        *logger.Logger
	serverId   values.ServerId
	serverType models.ServerType
	roleMap    map[int64]*role.LangRole
	lock       *sync.RWMutex
	roleCnt    values.Integer
}

func NewService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
) *Service {
	svc := routine_limit_service.NewRoutineLimitService(urls, log, serverId, serverType, true, false, eventlocal.CreateEventLocal(true))
	s := &Service{
		svc:        svc,
		log:        log,
		serverId:   serverId,
		serverType: serverType,
		lock:       &sync.RWMutex{},
	}
	return s
}

func (svc *Service) Serve() {
	svc.load()
	svc.Router()
	svc.svc.Start(func(event interface{}) {
		svc.log.Warn("unknown event", zap.Any("event", event))
	}, true)
}

func (svc *Service) Stop() {
	svc.svc.Close()
}

func (svc *Service) Router() {
	role.NewRoleService(svc.serverId, svc.serverType, svc.svc, svc.roleMap, svc.roleCnt, svc.log, svc.lock).Router()
}

const (
	page    = 200 // 每次查询的数量
	refresh = time.Hour * 24
)

func (svc *Service) load() {
	svc.loadRoleDataForTest()
	ticker := time.NewTicker(refresh)
	select {
	case <-ticker.C:
		svc.loadRoleDataForTest()
		ticker.Reset(refresh)
	default:
	}
}

func (svc *Service) loadRoleData() {
	roleMap := map[int64]*role.LangRole{}
	c := redisclient.GetDefaultRedis()
	ctx := context.Background()
	expected := 0
	tempMap := make(map[values.RoleId]struct{})
	for _, key := range enum.RoleAllIdKey {
		index := uint64(0)
		for {
			var keys []string
			var err error
			keys, index, err = c.HScan(ctx, key, index, "*", page).Result()
			if err != nil {
				panic(err)
			}
			newKeyS := make([]orm.RedisInterface, 0, len(keys)/2)
			for i := 0; i < len(keys); i += 2 {
				m := &pbdao.Role{
					RoleId: keys[i],
				}
				newKeyS = append(newKeyS, m)
				expected++
			}

			notFound, e := orm.GetOrm(ctx).MGetPB(c, newKeyS...)
			if e != nil {
				panic(e)
			}
			notFoundMap := make(map[int]bool, len(notFound))
			for _, v := range notFound {
				notFoundMap[v] = true
			}

			for i := 0; i < len(newKeyS); i++ {
				if _, ok := notFoundMap[i]; ok {
					continue
				}
				temp := role.NewRole(newKeyS[i].(*pbdao.Role))
				// HScan 可能会返回重复的key
				if _, ok := tempMap[temp.RoleId]; ok {
					continue
				}
				tempMap[temp.RoleId] = struct{}{}
				if _, ok := roleMap[temp.Language]; !ok {
					roleMap[temp.Language] = &role.LangRole{
						RoleList: make([]role.Role, 0),
						Lock:     &sync.RWMutex{},
					}
				}
				roleMap[temp.Language].RoleList = append(roleMap[temp.Language].RoleList, role.NewRole(newKeyS[i].(*pbdao.Role)))
				svc.roleCnt++
			}
			if index == 0 || len(keys) != page {
				break
			}
		}
	}

	svc.log.Info("load role data finish")
	for k, v := range roleMap {
		svc.log.Info("language", zap.Int(string(k), len(v.RoleList)))
	}
	svc.lock.Lock()
	svc.roleMap = roleMap
	svc.lock.Unlock()
}

func (svc *Service) loadRoleDataForTest() {
	svc.roleMap = map[int64]*role.LangRole{}
	svc.roleMap[1] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	svc.roleMap[2] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	svc.roleMap[3] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	svc.roleMap[4] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	svc.roleMap[5] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	for i := 0; i < 2000000; i++ {
		svc.roleMap[1].RoleList = append(svc.roleMap[1].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 1,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
		svc.roleCnt++
	}
	for i := 0; i < 2000000; i++ {
		svc.roleMap[2].RoleList = append(svc.roleMap[2].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 2,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
		svc.roleCnt++
	}
	for i := 0; i < 2000000; i++ {
		svc.roleMap[3].RoleList = append(svc.roleMap[3].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 3,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
		svc.roleCnt++
	}
	for i := 0; i < 2000000; i++ {
		svc.roleMap[4].RoleList = append(svc.roleMap[4].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 4,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
		svc.roleCnt++
	}
	for i := 0; i < 2000000; i++ {
		svc.roleMap[5].RoleList = append(svc.roleMap[5].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 5,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
		svc.roleCnt++
	}
}
