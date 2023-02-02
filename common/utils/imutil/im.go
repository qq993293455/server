package imutil

import (
	"context"
	"strconv"

	"coin-server/common/im"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/values/enum/Notice"

	jsoniter "github.com/json-iterator/go"
)

const (
	ExtraRoleLvKey          = "role_lv"           // 角色等级
	ExtraRoleAvatarIdKey    = "role_avatar_id"    // 角色头像
	ExtraRoleAvatarFrameKey = "role_avatar_frame" // 角色头像框

	ExtraShareEquipHeroStarKey = "share_equip_hero_star" // 分享装备英雄装备槽星级
)

func getIMRoleInfoExtra(role *daopb.Role) map[string]interface{} {
	extra := make(map[string]interface{})
	extra[ExtraRoleLvKey] = role.Level
	extra[ExtraRoleAvatarIdKey] = role.AvatarId
	extra[ExtraRoleAvatarFrameKey] = role.AvatarFrame
	return extra
}

func GetIMRoleInfoExtra(role *daopb.Role) string {
	extra := getIMRoleInfoExtra(role)
	_extra, _ := jsoniter.MarshalToString(extra)
	return _extra
}

func GetShareEquipExtra(role *daopb.Role, star string) string {
	extra := getIMRoleInfoExtra(role)
	extra[ExtraShareEquipHeroStarKey] = star
	_extra, _ := jsoniter.MarshalToString(extra)
	return _extra
}

type ContentNotice struct {
	NoticeID int64 `json:"notice_id"`
	Args     []any `json:"args"`
}

// GenNoticeContent 生成跑马灯content
func GenNoticeContent(id Notice.Enum, args ...any) string {
	content, err := jsoniter.MarshalToString(&ContentNotice{NoticeID: id, Args: args})
	if err != nil {
		panic(err)
	}
	return content
}

// SendNotice 发跑马灯消息
func SendNotice(ctx context.Context, parseType int, sendChat bool, noticeId Notice.Enum, args ...any) (err error) {
	msg := &im.Message{
		Type:       im.MsgTypeBroadcast,
		RoleID:     "admin",
		RoleName:   "admin",
		Content:    GenNoticeContent(noticeId, args...),
		ParseType:  parseType,
		IsMarquee:  true,
		IsVolatile: !sendChat,
	}
	return im.DefaultClient.SendMessage(ctx, msg)
}

// SendNoticeByBattleLine 按战斗分线发跑马灯消息
func SendNoticeByBattleLine(ctx context.Context, line string, parseType int, sendChat bool, noticeId Notice.Enum, args ...any) (err error) {
	msg := &im.Message{
		Type:       im.MsgTypeRoom,
		RoleID:     "admin",
		RoleName:   "admin",
		TargetID:   line,
		Content:    GenNoticeContent(noticeId, args...),
		ParseType:  parseType,
		IsMarquee:  true,
		IsVolatile: !sendChat,
	}
	return im.DefaultClient.SendMessage(ctx, msg)
}

// BattleLineRoom 战斗分线频道
func BattleLineRoom(battleServerId int64) string {
	return "battle_line_" + strconv.Itoa(int(battleServerId))
}
