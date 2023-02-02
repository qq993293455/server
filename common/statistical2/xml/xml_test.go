package xml

import (
	"fmt"
	"testing"

	"coin-server/common/statistical/models"
	"coin-server/common/utils/test"
)

var tm *test.ServerTestMain

func TestMain(m *testing.M) {
	fmt.Println("test start")
	tm = test.NewServerTestMain()
	m.Run()
}

func Test_genXML(t *testing.T) {
	tm.NewAstAndReq(t)
	genTablesXML(models.TopicModelMap)
}
