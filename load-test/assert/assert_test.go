package assert

import (
	"testing"

	"coin-server/common/errmsg"
	"coin-server/load-test/core"

	testifyAst "github.com/stretchr/testify/assert"
)

func TestAssert(t *testing.T) {
	testifyAst.NotPanicsf(t, func() {
		ctx := &core.RoleContext{
			UserId: "test-userId",
			RoleId: "test-roleId",
		}
		Equal(ctx, 1, 1)
		NotEqual(ctx, 1, 2)
		Nil(ctx, nil)
		NotNil(ctx, "dfdfdfd")
		True(ctx, true)
		False(ctx, false)
		Empty(ctx, "")
		Empty(ctx, nil)
		Empty(ctx, []int{})
		Empty(ctx, map[int]int{})
		Empty(ctx, struct{}{})
		NotEmpty(ctx, "a")
		Error(ctx, errmsg.NewErrAchievementAlreadyCollect())
		ErrorIs(ctx, errmsg.NewErrAchievementAlreadyCollect(), errmsg.NewErrAchievementAlreadyCollect())
	}, "not panic")
}
