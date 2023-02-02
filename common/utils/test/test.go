package test

import (
	"testing"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/proto/models"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ServerTestMain struct {
	testVersion string
	Reader      *rulemodel.Reader
	Ctx         *ctx.Context
	Ast         *assert.Assertions
	Req         *require.Assertions
	Ctrl        *gomock.Controller
}

func NewServerTestMain() *ServerTestMain {
	rule.LoadRuleByFile()
	tm := &ServerTestMain{
		testVersion: "",
		Reader:      nil,
		Ctx:         nil,
		Ast:         nil,
		Req:         nil,
		Ctrl:        nil,
	}

	tm.Ctx = &ctx.Context{}
	tm.Ctx.ServerHeader = &models.ServerHeader{
		StartTime:         time.Now().UnixMilli(),
		RoleId:            "",
		ServerId:          0,
		ServerType:        0,
		RuleVersion:       "",
		TraceId:           "",
		GateId:            0,
		UserId:            "",
		InServerId:        0,
		BattleServerId:    0,
		BattleMapId:       0,
		StateLessServerId: 0,
	}
	reader := rule.MustGetReader(tm.Ctx)
	tm.Reader = reader
	return tm
}

func (tm *ServerTestMain) NewAstAndReq(t *testing.T) {
	tm.Ast = assert.New(t)
	tm.Req = require.New(t)
	tm.Ctrl = gomock.NewController(t)
	defer tm.Ctrl.Finish()
}
