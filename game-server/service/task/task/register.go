package task

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
	event2 "coin-server/common/values/event"
)

type Register struct {
	handlers map[models.TaskType]map[values.Integer]map[values.Integer][]*event2.Handler
}

func NewRegister() *Register {
	instance := &Register{
		handlers: map[models.TaskType]map[values.Integer]map[values.Integer][]*event2.Handler{},
	}
	return instance
}

func (r *Register) RegisterHandler(taskType models.TaskType, targetId, targetCnt values.Integer, handler event2.CondHandler, args any) {
	if r.handlers[taskType] == nil {
		r.handlers[taskType] = map[values.Integer]map[values.Integer][]*event2.Handler{}
	}
	if r.handlers[taskType][targetId] == nil {
		r.handlers[taskType][targetId] = map[values.Integer][]*event2.Handler{}
	}
	if r.handlers[taskType][targetId][targetCnt] == nil {
		r.handlers[taskType][targetId][targetCnt] = make([]*event2.Handler, 0)
	}
	r.handlers[taskType][targetId][targetCnt] = append(r.handlers[taskType][targetId][targetCnt], &event2.Handler{
		Args: args,
		H:    handler,
	})
}
