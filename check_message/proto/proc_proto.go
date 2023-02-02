package proto

import (
	"coin-server/common/msgcreate"
	_ "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	_ "coin-server/common/proto/roguelike_match"
	pb "coin-server/common/proto/service"
	"fmt"

	"coin-server/check_message/assert"
	"coin-server/check_message/core"
	"coin-server/check_message/user"
)

type ProtoPackage struct {
	protoPkg SourceProto
}

func GetProtoPackage() *ProtoPackage {
	return &ProtoPackage{}
}

func (this_ *ProtoPackage) InitProto(path string) {
	this_.protoPkg.ParseProtoFiles(path)
}

func (this_ *ProtoPackage) InstanceProto(addr string, serverId int64) {
	// for name := range this_.protoPkg.MessageMap {
	// 	fmt.Println(name)
	// 	req := msgcreate.NewMessage(name)
	// 	fmt.Println(proto.MessageName(req))
	// }
	// for name, pkgT =:range this_.protoPkg.MessageMap{
	// req := msgcreate.NewMessage(msgName)
	// }

	// goroutine创建时先获得一个uid
	uid, err := core.GetUserId()
	if err != nil {
		panic("init userId fail," + err.Error())
	}
	ctx := core.NewRoleContext(uid)
	// 初始化该用户的网络连接
	fmt.Println("connect : "+addr, " server id : ", serverId)
	ctx.IConnect = core.NewTcpConn(ctx, addr)

	// 执行登录操作
	err1 := user.Login(ctx, serverId)
	if err1 != nil {
		panic("login fail !uid:" + uid + " " + err1.Error() + " ")
	}

	//UnlockSystem(ctx)
	fmt.Println("check start")
	for name := range this_.protoPkg.MessageMap {
		req := msgcreate.NewMessage(name)
		_, _, errReq := ctx.Request(req)
		if errReq != nil {
			if errReq.ErrMsg != "gm_handler_closed" {
				fmt.Println("no proc message: "+name, errReq.String(), errReq.ErrMsg, errReq.ErrCode)
			}
		} else {
			fmt.Println("no middler : " + name)
		}
	}

	fmt.Println("check over")
}

func UnlockSystem(ctx *core.RoleContext) {
	req := &pb.SystemUnlock_CheatUnlockSystemRequest{
		SystemId: models.SystemType_SystemArena,
	}
	_, _, err := ctx.Request(req)
	assert.Nil(ctx, err)
}
