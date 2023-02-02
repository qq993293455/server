package role

import (
	"sync"

	"coin-server/common/proto/dao"
)

type Role struct {
	RoleId   string
	Nickname string
	Power    int64
	Level    int64
	Language int64
	Login    int64
	Logout   int64
}

type LangRole struct {
	RoleList []Role
	Lock     *sync.RWMutex
}

func NewRole(role *dao.Role) Role {
	return Role{
		RoleId:   role.RoleId,
		Nickname: role.Nickname,
		Power:    role.Power,
		Level:    role.Level,
		Language: role.Language,
		Login:    role.Login,
		Logout:   role.Logout,
	}
}
