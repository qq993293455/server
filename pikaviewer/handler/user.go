package handler

import (
	"time"

	"coin-server/common/utils"
	"coin-server/pikaviewer/model"
	utils2 "coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

type User struct{}

func (h *User) Login(username, password string) (gin.H, error) {
	user := model.NewUser()
	ok, err := user.GetByName(username)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, utils2.NewDefaultErrorWithMsg("用户不存在")
	}
	pwd := utils.MD5String(password + utils2.PwdMD5Key)
	if user.Password != pwd {
		return nil, utils2.NewDefaultErrorWithMsg("密码错误")
	}
	duration := time.Hour * 12
	token, err := utils2.GenToken(user.Id, user.Role, duration)
	if err != nil {
		return nil, err
	}
	ret := gin.H{
		"uid":      user.Role,
		"username": user.Username,
		"role":     user.Role,
		"token":    token,
		"expireAt": time.Now().Add(duration).UnixMilli(),
	}
	return ret, nil
}
