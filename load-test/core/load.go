package core

type ILoadTestModule interface {
	// 每个单次压测入口
	Process(ctx *RoleContext)
}

type ModuleFactory func(ctx *RoleContext) ILoadTestModule
