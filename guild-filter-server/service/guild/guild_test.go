package guild

import (
	"testing"

	"coin-server/common/utils/test"
)

var tm *test.ServerTestMain

func TestMain(m *testing.M) {
	tm = test.NewServerTestMain()
	m.Run()
}
