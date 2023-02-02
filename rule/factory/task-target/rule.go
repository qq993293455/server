package tasktarget

import (
	"errors"

	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type Param struct {
	TaskType models.TaskType
	Target   values.Integer
	Count    values.Integer
	// 自定义参数
	P1 values.Integer
}

func ParseParam(params []values.Integer) *Param {
	if len(params) < 3 {
		return nil
	}
	switch len(params) {
	case 3:
		return &Param{
			TaskType: models.TaskType(params[0]),
			Target:   params[1],
			Count:    params[2],
		}
	default:
		return &Param{
			TaskType: models.TaskType(params[0]),
			Target:   params[1],
			Count:    params[2],
			P1:       params[3],
		}
	}
}

func MustParseParam(params []values.Integer) *Param {
	if len(params) < 3 {
		panic(errors.New("task target params len < 3"))
	}
	switch len(params) {
	case 3:
		return &Param{
			TaskType: models.TaskType(params[0]),
			Target:   params[1],
			Count:    params[2],
		}
	default:
		return &Param{
			TaskType: models.TaskType(params[0]),
			Target:   params[1],
			Count:    params[2],
			P1:       params[3],
		}
	}
}
