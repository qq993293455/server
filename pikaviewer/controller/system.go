package controller

import (
	"os"

	"coin-server/common/values/env"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

type System struct {
}

func (c *System) Name(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	resp.Send(os.Getenv(env.APP_NAME))
}
