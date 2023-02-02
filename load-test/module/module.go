package module

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"

	"coin-server/common/values/env"
	"coin-server/load-test/core"
	"coin-server/load-test/module/user"

	"gitlab.cdl5.org/chengdu-l5/boomer"
)

var ts = boomer.NewWeighingTaskSet()
var allFactories = map[string]core.ModuleFactory{}

func Register(f core.ModuleFactory, wait int) {
	name := getPackageName(f)
	if _, ok := allFactories[name]; ok {
		panic("module: " + name + " already registered")
	}
	allFactories[name] = f

	if wait <= 0 {
		wait = 100
	}
	task := &boomer.Task{
		Name:   name,
		Weight: wait,
		FnArgs: func(args interface{}) {
			ctx := args.(*core.RoleContext)
			defer func() {
				// don't panic
				err := recover()
				if err != nil {
					stackTrace := debug.Stack()
					errMsg := fmt.Sprintf("%v", err)
					os.Stderr.Write([]byte("userId:" + ctx.UserId + " roleId:" + ctx.RoleId + " "))
					os.Stderr.Write([]byte(errMsg))
					os.Stderr.Write([]byte("\n"))
					os.Stderr.Write(stackTrace)
					return
				}
			}()
			ctx.Context = context.Background()
			s := ctx.GetModuleService(name)
			s.Process(ctx)
		},
	}
	ts.AddTask(task)
}

// 比如 coin-server/load-test/module/user.New  返回 user
func getPackageName(f core.ModuleFactory) string {
	str := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	list := strings.Split(str, "/")
	str2 := list[len(list)-1]
	list2 := strings.Split(str2, ".")
	return list2[0]
}

func GetModuleTasks() *boomer.Task {
	spawnFn := func() interface{} {
		// goroutine创建时先获得一个uid
		uid, err := core.GetUserId()
		if err != nil {
			panic("init userId fail," + err.Error())
		}
		ctx := core.NewRoleContext(uid)
		// 初始化该用户的网络连接
		ctx.IConnect = core.NewTcpConn(ctx, os.Getenv(env.LOCUST_TARGET_SERVER_ADDR))
		// 初始化该用户独有的所有module实例
		for name, f := range allFactories {
			s := f(ctx)
			ctx.SetModuleService(name, s)
		}
		// 执行登录操作
		err1 := user.Login(ctx)
		if err1 != nil {
			panic("login fail !uid:" + uid + " " + err1.Error() + " ")
		}
		// 开始后续业务逻辑
		return ctx
	}
	loopFn := func(args interface{}) {
		ts.RunWithArgs(args)
	}
	task := &boomer.Task{
		Name:    "TaskSet",
		FnArgs:  loopFn,
		SpawnFn: spawnFn,
	}
	return task
}
