package rank

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"coin-server/common/proto/dao"
	"coin-server/common/utils/test"
)

var tm *test.ServerTestMain

func TestMain(m *testing.M) {
	tm = test.NewServerTestMain()
	m.Run()
}

func Test_handleTopRank(t *testing.T) {
	tm.NewAstAndReq(t)
	svc := NewRankService(0, 0, nil, nil)
	data := make(map[string]*dao.TopRankItem, 0)
	for i := 0; i < 5; i++ {
		data[strconv.Itoa(i)] = &dao.TopRankItem{
			CombatValue: rand.Int63n(10000),
			CreatedAt:   rand.Int63n(10000),
		}
	}
	ret := svc.handleTopRank(data, 2)
	fmt.Println(ret)
}
