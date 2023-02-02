package entryskill

import (
	"fmt"

	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type Param struct {
	SkillType models.EntrySkillType
	// 自定义参数
	P1 values.Integer
	P2 values.Integer
}

func ParseTyp(params []values.Integer) *Param {
	if len(params) < 2 {
		return nil
	}
	switch len(params) {
	case 3:
		return &Param{
			SkillType: models.EntrySkillType(params[0]),
			P1:        params[1],
			P2:        params[2],
		}
	default:
		return &Param{
			SkillType: models.EntrySkillType(params[0]),
			P1:        params[1],
		}
	}
}

func MustParseTyp(params []values.Integer) *Param {
	ret := ParseTyp(params)
	if ret == nil {
		panic(fmt.Sprintf("skillType Typ error: %v", params))
	}
	return ret
}
