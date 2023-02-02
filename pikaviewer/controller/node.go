package controller

import (
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

type Nodes struct {
}

func NewNodes() *Nodes {
	return &Nodes{}
}

func (m *Nodes) ModuleModes(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	resp.Send(handler.NewNodes().ModuleNodes())
}

func (m *Nodes) PikaNodes(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	resp.Send(handler.NewNodes().PikaNodes())
}
