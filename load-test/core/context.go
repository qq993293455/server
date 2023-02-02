package core

import (
	"context"

	"coin-server/common/logger"
	"coin-server/common/values"
)

type RoleContext struct {
	context.Context
	IConnect
	*logger.Logger
	UserId string
	RoleId values.RoleId

	AllModules map[string]ILoadTestModule
}

func NewRoleContext(userId string) *RoleContext {
	ctx := &RoleContext{
		Context:    context.Background(),
		UserId:     userId,
		Logger:     logger.DefaultLogger,
		AllModules: make(map[string]ILoadTestModule),
	}
	return ctx
}

func (ctx *RoleContext) GetModuleService(name string) ILoadTestModule {
	if v, ok := ctx.AllModules[name]; ok {
		return v
	}
	return nil
}

func (ctx *RoleContext) SetModuleService(name string, v ILoadTestModule) {
	ctx.AllModules[name] = v
}

func (ctx *RoleContext) RecordSuccess(name string, responseTime values.Integer, responseLength values.Integer) {
	if GlobalBoomer != nil {
		GlobalBoomer.RecordSuccess("tcp", name, responseTime, responseLength)
	}
	return
}

func (ctx *RoleContext) RecordFailure(name string, responseTime values.Integer, exception string) {
	if GlobalBoomer != nil {
		GlobalBoomer.RecordFailure("tcp", name, responseTime, exception)
	}
	return
}
