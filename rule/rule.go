package rule

import (
	"errors"

	"coin-server/common/consulkv"
	"coin-server/common/ctx"
	"coin-server/common/logger"

	"go.uber.org/zap"
)

var addr string

func Load(cnf *consulkv.Config) {
	var ok bool
	addr, ok = cnf.GetString("newhttp/rule")
	if !ok {
		panic(errors.New("can't find ruleCenter addr in consul"))
	}
	LoadRule()
	r := MustGetReader(ctx.GetContext())
	logger.DefaultLogger.Info("rule load success", zap.String("version", r.Version))
}
