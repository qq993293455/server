package selector

import (
	"coin-server/common/consulkv"
	"coin-server/pikaviewer/selector/implement"

	"github.com/gin-gonic/gin"
)

var Gateway Selector

type Selector interface {
	Get(ctx *gin.Context) string
}

func Init(config *consulkv.Config) {
	Gateway = implement.New(config)
}
