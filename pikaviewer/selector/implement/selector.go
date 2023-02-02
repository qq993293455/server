package implement

import (
	"coin-server/common/consulkv"
	"coin-server/common/ipregion"
	"coin-server/common/logger"
	"coin-server/common/utils"
	"go.uber.org/zap"

	"stathat.com/c/consistent"

	"github.com/gin-gonic/gin"
)

type gateway struct {
	US map[string]int64 `json:"us"`
	SG map[string]int64 `json:"sg"`
}

func New(config *consulkv.Config) *gateway {
	cfg := &gateway{}
	utils.Must(config.Unmarshal("gateway/list", cfg))
	logger.DefaultLogger.Debug("read gateway config from consul", zap.Any("gateway/list", cfg))
	return cfg
}

func (g *gateway) Get(ctx *gin.Context) string {
	var data map[string]int64
	if ipregion.IsNorthAmerica(ctx.ClientIP()) {
		data = g.US
	} else {
		data = g.SG
	}
	c := consistent.New()
	var defaultVal string
	for item := range data {
		c.Add(item)
		if defaultVal == "" {
			defaultVal = item
		}
	}
	val, err := c.Get(ctx.Query("device"))
	if err != nil {
		val = defaultVal
	}
	return val
}
