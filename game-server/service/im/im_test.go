package im

import (
	"fmt"
	"testing"
	"unsafe"

	"coin-server/common/proto/models"
	"coin-server/common/utils/test"

	json "github.com/json-iterator/go"
)

var tm *test.ServerTestMain

func TestMain(m *testing.M) {
	tm = test.NewServerTestMain()
	m.Run()
}

func Test_Marshal(t *testing.T) {
	//c936g2tjvqn49e25f0vg
	//c936g2tjvqn49e25f0rg
	equipList := make([]*Equipment, 0)
	equip := &models.Equipment{
		EquipId: "aaa",
		ItemId:  99,
		Level:   99,
		Affix: []*models.Affix{
			{
				AffixId:    1,
				Quality:    1,
				AffixValue: 1,
				SkillId:    1,
				Active:     false,
				AttrId:     1,
				Bonus: map[int64]int64{
					1: 1,
				},
			},
		},
		HeroId: 0,
	}
	e := (*Equipment)(unsafe.Pointer(equip))
	equipList = append(equipList, e)

	content, err1 := json.MarshalToString(equipList)
	fmt.Println(content, err1)
}
