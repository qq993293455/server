package mail

import (
	"coin-server/common/proto/models"
	pbSvr "coin-server/common/proto/service"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
)

// 每个用户对应单独的一个service实例
type service struct {
	unread map[string]*models.Mail
	read   map[string]*models.Mail
}

func New(ctx *core.RoleContext) core.ILoadTestModule {
	return &service{}
}

func (s *service) Process(ctx *core.RoleContext) {
	s.reset(ctx)
	s.cheatAddMail(ctx)
	s.getMailList(ctx)
	s.readAllMail(ctx)
	s.delAllMail(ctx)
}

func (s *service) reset(ctx *core.RoleContext) {
	s.unread, s.read = nil, nil
}

func (s *service) cheatAddMail(ctx *core.RoleContext) {
	req := &pbSvr.Mail_CheatAddMailRequest{
		TextId: 10001, // TODO
		Args:   []string{"test-args-1", "test-args-2"},
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Mail_CheatAddMailResponse)
}

func (s *service) getMailList(ctx *core.RoleContext) {
	req := &pbSvr.Mail_MailListRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	result := res.(*pbSvr.Mail_MailListResponse)
	s.unread = result.Unread
	s.read = result.Read
}

func (s *service) readAllMail(ctx *core.RoleContext) {
	req := &pbSvr.Mail_ReadAllMailRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Mail_ReadAllMailResponse)
}

func (s *service) delAllMail(ctx *core.RoleContext) {
	req := &pbSvr.Mail_DeleteAllMailRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Mail_DeleteAllMailResponse)
}
