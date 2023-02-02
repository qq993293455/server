package user

import (
	"unsafe"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
)

const addExpLock = "lock:addExp:"

func lock(ctx *ctx.Context, userId string) *errmsg.ErrMsg {
	return ctx.DRLock(redisclient.GetLocker(), "lock:user:"+userId)
}

func RoleModel2Dao(roleModel *models.Role) *dao.Role {
	return (*dao.Role)(roleModel)
}

func RoleDao2Model(roleDao *dao.Role) *models.Role {
	return (*models.Role)(roleDao)
}

func RoleModels2Dao(roleModel []*models.Role) []*dao.Role {
	return *(*[]*dao.Role)(unsafe.Pointer(&roleModel))
}

func RoleDao2Models(roleDao []*dao.Role) []*models.Role {
	return *(*[]*models.Role)(unsafe.Pointer(&roleDao))
}

func RoleDao2SimpleModels(roleDao []*dao.Role) []*models.RoleSimple {
	res := make([]*models.RoleSimple, 0, len(roleDao))
	for _, role := range roleDao {
		res = append(res, &models.RoleSimple{
			RoleId:      role.RoleId,
			Nickname:    role.Nickname,
			AvatarId:    role.AvatarId,
			AvatarFrame: role.AvatarFrame,
			Level:       role.Level,
			Power:       role.Power,
		})
	}
	return res
}

func RoleAttrModel2Dao(roleAttrModel *models.RoleAttr) *dao.RoleAttr {
	return (*dao.RoleAttr)(roleAttrModel)
}

func RoleAttrDao2Model(roleAttrDao *dao.RoleAttr) *models.RoleAttr {
	return (*models.RoleAttr)(roleAttrDao)
}

func RoleAttrModels2Dao(roleAttrModel []*models.RoleAttr) []*dao.RoleAttr {
	return *(*[]*dao.RoleAttr)(unsafe.Pointer(&roleAttrModel))
}

func RoleAttrDao2Models(roleAttrDao []*dao.RoleAttr) []*models.RoleAttr {
	return *(*[]*models.RoleAttr)(unsafe.Pointer(&roleAttrDao))
}
