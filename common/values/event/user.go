package event

import "coin-server/common/values"

var UserLogin values.EventName = "UserLoginIn"

type UserLoginData struct {
	RoleId values.RoleId
}
