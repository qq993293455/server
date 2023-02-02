package controller

import (
	utils2 "coin-server/common/utils"
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

type Query struct {
}

func NewQuery() *Query {
	return &Query{}
}

func (q *Query) Do(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	qh := handler.NewQuery()
	if err := ctx.ShouldBindJSON(qh); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("参数有误"))
		return
	}
	if qh.Way == "instance" && qh.Instance == "" {
		resp.Send(utils.NewDefaultErrorWithMsg("参数有误"))
		return
	}
	data, err := qh.Do()
	if !utils2.IsNil(err) && err != nil {
		resp.Send(err)
		return
	}
	resp.Send(data)
}
