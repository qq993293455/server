package system_unlock

//
// import (
// 	"strings"
// 	"sync"
//
// 	"coin-server/common/ctx"
// 	"coin-server/common/errmsg"
// 	"coin-server/common/handler"
// 	"coin-server/common/proto/models"
// 	"coin-server/common/utils"
// 	"coin-server/common/values"
// 	"coin-server/game-server/service/system-unlock/dao"
// 	"coin-server/rule"
//
// 	"github.com/gogo/protobuf/proto"
// )
//
// const segment = 1 << 4
//
// var cache = NewCache()
//
// type Cache struct {
// 	unlock [segment]map[values.RoleId]map[models.SystemType]bool
// 	locker [segment]*sync.RWMutex
// }
//
// func NewCache() *Cache {
// 	c := &Cache{
// 		unlock: [segment]map[values.RoleId]map[models.SystemType]bool{},
// 		locker: [segment]*sync.RWMutex{},
// 	}
// 	for i := 0; i < segment; i++ {
// 		c.unlock[i] = map[values.RoleId]map[models.SystemType]bool{}
// 		c.locker[i] = &sync.RWMutex{}
// 	}
// 	return c
// }
//
// func GetSysUnlockCache() *Cache {
// 	return cache
// }
//
// func (c *Cache) init(ctx *ctx.Context, unlock map[values.Integer]bool) {
// 	slot := hashRoleId(ctx.RoleId)
// 	m := map[models.SystemType]bool{}
// 	for k, v := range unlock {
// 		m[models.SystemType(k)] = v
// 	}
// 	c.locker[slot].Lock()
// 	defer c.locker[slot].Unlock()
// 	delete(c.unlock[slot], ctx.RoleId)
// 	c.unlock[slot][ctx.RoleId] = m
// }
//
// func (c *Cache) isUnlock(ctx *ctx.Context, typ models.SystemType) bool {
// 	slot := hashRoleId(ctx.RoleId)
// 	c.locker[slot].RLock()
// 	defer c.locker[slot].RUnlock()
// 	rl, ok := c.unlock[slot][ctx.RoleId]
// 	if !ok {
// 		return true
// 	}
// 	unlock, ok1 := rl[typ]
// 	if !ok1 {
// 		return true
// 	}
// 	return unlock
// }
//
// func (c *Cache) SetUnlock(ctx *ctx.Context, typ models.SystemType) {
// 	slot := hashRoleId(ctx.RoleId)
// 	c.locker[slot].Lock()
// 	defer c.locker[slot].Unlock()
// 	rl, ok := c.unlock[slot][ctx.RoleId]
// 	if !ok {
// 		rl = map[models.SystemType]bool{}
// 		c.unlock[slot][ctx.RoleId] = rl
// 	}
// 	rl[typ] = true
// 	return
// }
//
// func hashRoleId(roleId values.RoleId) int {
// 	return int(utils.Base34DecodeString(roleId) & (segment - 1))
// }
//
// func SystemChecker(next handler.HandleFunc) handler.HandleFunc {
// 	return func(ctx *ctx.Context) *errmsg.ErrMsg {
// 		var reqMsgName string
// 		if ctx.ServerType == models.ServerType_GatewayStdTcp && ctx.Req != nil {
// 			reqMsgName = proto.MessageName(ctx.Req)
// 			str := strings.Split(reqMsgName, ".")
// 			if len(str) >= 3 {
// 				sys := str[1]
// 				r := rule.MustGetReader(ctx)
// 				typ, ok := r.System.GetSysTypeByName(sys)
// 				if ok && !cache.isUnlock(ctx, typ) {
// 					return errmsg.NewErrSystemLock()
// 				}
// 			}
// 		}
// 		return next(ctx)
// 	}
// }
