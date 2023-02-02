package controller

import (
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var userHandler = new(handler.User)

type User struct {
	LoginForm struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
}

func (c *User) Login(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &c.LoginForm
	if err := ctx.ShouldBindJSON(req); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("请输入用户名和密码"))
		return
	}
	user, err := userHandler.Login(req.Username, req.Password)
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(user)
}
