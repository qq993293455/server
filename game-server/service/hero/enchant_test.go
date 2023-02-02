package hero

import (
	"fmt"
	"testing"

	"coin-server/common/values"
	"coin-server/game-server/module"
)

func Test_handleProb(t *testing.T) {
	tm.NewAstAndReq(t)
	svc := NewHeroService(0, 0, nil, &module.Module{}, nil)
	weightMap := map[values.Integer]values.Integer{1: 1000, 2: 1000, 3: 1000, 4: 1000, 5: 1000, 6: 1000}
	fmt.Printf("%+v\n", svc.handleProb(weightMap))
}

func Test_abc(t *testing.T) {
	data := map[int64]int64{101: 100}
	ud := &UpgradeData{
		Cost:     nil,
		Upgraded: nil,
		Count:    0,
	}
	mapTest(ud)
	fmt.Println(data)
}

func mapTest(data *UpgradeData) {
	data.Cost[101] += 100
	data.Count++
}
