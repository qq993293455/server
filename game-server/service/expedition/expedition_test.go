package expedition

import (
	"testing"

	"coin-server/common/utils/test"
	"coin-server/game-server/module"
)

var tm *test.ServerTestMain

func TestMain(m *testing.M) {
	tm = test.NewServerTestMain()
	m.Run()
}

func Test_getSlotCount(t *testing.T) {
	tm.NewAstAndReq(t)
	svc := NewExpeditionService(0, 0, nil, &module.Module{}, nil)
	svc.getSlotCount(tm.Ctx)
}
