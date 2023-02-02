package guild

import (
	"os"

	"coin-server/common/errmsg"
	userSvrPb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/values/env"
	"coin-server/load-test/core"
	defaultEnv "coin-server/load-test/env"

	"github.com/gogo/protobuf/proto"
)

type testUser struct {
	ctx *core.RoleContext
}

func newTestUser() *testUser {
	t := &testUser{}
	uid, err := core.GetUserId()
	if err != nil {
		panic("init userId fail," + err.Error())
	}
	t.ctx = core.NewRoleContext(uid)
	// 初始化该用户的网络连接
	t.ctx.IConnect = core.NewTcpConn(t.ctx, os.Getenv(env.LOCUST_TARGET_SERVER_ADDR))
	// 执行登录操作
	err1 := t.login()
	if err1 != nil {
		panic("login fail !uid:" + uid + " " + err1.Error() + " ")
	}
	t.registerHeartbeat()
	return t
}

func (t *testUser) login() *errmsg.ErrMsg {
	req := &userSvrPb.User_RoleLoginRequest{
		UserId:       t.ctx.UserId,
		ServerId:     defaultEnv.GetTargetServerId(),
		LessServerId: defaultEnv.GetTargetLessServerId(),
	}
	_, res, err := t.ctx.Request(req)
	if err != nil {
		panic(err)
	}
	response := res.(*userSvrPb.User_RoleLoginResponse)
	//ctx.RoleInfo = response.Date.RoleInfo
	t.ctx.RoleId = response.RoleId
	return nil
}

func (t *testUser) registerHeartbeat() {
	ping := &models.PING{}
	t.ctx.RegisterPush(ping.XXX_MessageName(), func(ctx *core.RoleContext, msg proto.Message) *errmsg.ErrMsg {
		_ = msg.(*models.PING)
		_ = ctx.AsyncSend(&models.PONG{})
		return nil
	})
}
