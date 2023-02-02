package service

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/recommend"
	"coin-server/common/values"
	"coin-server/recommend-server/service/role"
)

var roleMap = map[int64]*role.LangRole{}
var lock = &sync.RWMutex{}
var roleCnt = 10000000

func loadRoleData() {
	roleMap = map[int64]*role.LangRole{}
	roleMap[1] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	roleMap[2] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	roleMap[3] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	roleMap[4] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	roleMap[5] = &role.LangRole{
		RoleList: make([]role.Role, 0),
		Lock:     &sync.RWMutex{},
	}
	for i := 0; i < 2000000; i++ {
		roleMap[1].RoleList = append(roleMap[1].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 1,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
	}
	for i := 0; i < 2000000; i++ {
		roleMap[2].RoleList = append(roleMap[2].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 2,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
	}
	for i := 0; i < 2000000; i++ {
		roleMap[3].RoleList = append(roleMap[3].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 3,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
	}
	for i := 0; i < 2000000; i++ {
		roleMap[4].RoleList = append(roleMap[4].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 4,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
	}
	for i := 0; i < 2000000; i++ {
		roleMap[5].RoleList = append(roleMap[5].RoleList, role.Role{
			RoleId:   "ABCDE",
			Nickname: "ABCDE",
			Power:    1000,
			Level:    1,
			Language: 5,
			Login:    time.Now().UnixMilli(),
			Logout:   0,
		})
	}
	return
}

func Add(roleId values.RoleId, lang int64) {
	lock.Lock()
	roleMap[lang].RoleList = append(roleMap[lang].RoleList, role.Role{RoleId: roleId})
	lock.Unlock()
}

func Recommend(roleId values.RoleId, lang int64) (*recommend.Recommend_RecommendResponse, *errmsg.ErrMsg) {
	var list []string
	num := 10
	if lang == 0 {
		return nil, nil
	}

	// 总人数小于等于推荐的人数，推荐全部人
	lock.RLock()
	if int(roleCnt) <= num {
		if int(roleCnt) == 0 {
			return nil, nil
		}
		for _, v := range roleMap {
			for idx := range v.RoleList {
				if roleId == v.RoleList[idx].RoleId {
					continue
				}
				list = append(list, v.RoleList[idx].RoleId)
			}
		}
		lock.RUnlock()
	} else {
		lock.RUnlock()
		list = doRecommendN(roleId, num, lang)
		// 本语言人数不足推荐人数, 推荐其他语言
		if len(list) < num {
			other := doRecommendOther(num-len(list), lang)
			list = append(list, other...)
		}
	}

	return &recommend.Recommend_RecommendResponse{
		Id: list,
	}, nil
}

func doRecommendN(roleId values.RoleId, num int, lang int64) []string {
	roles, ok := roleMap[lang]
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
			if roleId != roles.RoleList[i].RoleId {
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
		if roleId == roles.RoleList[i].RoleId {
			i--
			continue
		}
		list[i] = roles.RoleList[i].RoleId
		randMap[r] = struct{}{}
	}
	return list
}

// 随机推荐其他语种（公平）
func doRecommendOther(num int, lang int64) []string {
	list := make([]string, 0)
	lock.RLock()
	defer lock.RUnlock()
	for i := 0; i < num; i++ {
		r := rand.Intn(len(roleMap))
		n := 0
		for k, v := range roleMap {
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

func doRecommend(ctx *ctx.Context, lang int64) string {
	roles, ok := roleMap[lang]
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

func BenchmarkNewService(b *testing.B) {
	loadRoleData()
	b.ResetTimer()
	go func() {
		for {
			Add("BBBBB", 1)
			time.Sleep(time.Millisecond)
		}
	}()
	for i := 0; i < b.N; i++ {
		Recommend("AAAAA", 1)
	}
	//b.RunParallel(func(pb *testing.PB)  {
	//	for pb.Next() {
	//		Add("BBBBB", "CN")
	//		Recommend("AAAAA", "CN")
	//	}
	//})
}
