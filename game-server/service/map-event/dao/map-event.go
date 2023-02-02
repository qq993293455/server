package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetMapStory(ctx *ctx.Context, roleId values.RoleId) ([]*dao.MapStory, *errmsg.ErrMsg) {
	story := make([]*dao.MapStory, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getMapStoryKey(roleId), &story)
	if err != nil {
		return nil, err
	}
	return story, nil
}

func GetMapStoryById(ctx *ctx.Context, roleId values.RoleId, storyId values.StoryId) (*dao.MapStory, *errmsg.ErrMsg) {
	story := &dao.MapStory{StoryId: storyId}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getMapStoryKey(roleId), story)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return story, nil
}

func SaveMapStory(ctx *ctx.Context, roleId values.RoleId, story *dao.MapStory) *errmsg.ErrMsg {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getMapStoryKey(roleId), story)
	return nil
}

func SaveManyMapStory(ctx *ctx.Context, roleId values.RoleId, story []*dao.MapStory) {
	if len(story) == 0 {
		return
	}
	add := make([]orm.RedisInterface, len(story))
	for idx := range add {
		add[idx] = story[idx]
	}
	ctx.NewOrm().HMSetPB(redisclient.GetUserRedis(), getMapStoryKey(roleId), add)
	return
}

func GetMapEvent(ctx *ctx.Context, roleId values.RoleId) (*dao.MapEvent, *errmsg.ErrMsg) {
	event := &dao.MapEvent{RoleId: roleId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), event)
	if err != nil {
		return event, err
	}
	if !has {
		return nil, nil
	}
	return event, nil
}

func SaveMapEvent(ctx *ctx.Context, event *dao.MapEvent) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), event)
}

func GetStoryPiece(ctx *ctx.Context, roleId values.RoleId) (*dao.StoryPiece, *errmsg.ErrMsg) {
	story := &dao.StoryPiece{RoleId: roleId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), story)
	if err != nil {
		return story, err
	}
	if !has {
		return nil, nil
	}
	return story, nil
}

func SaveStoryPiece(ctx *ctx.Context, story *dao.StoryPiece) *errmsg.ErrMsg {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), story)
	return nil
}

func GetEventRefresh(ctx *ctx.Context, roleId values.RoleId) (*dao.MapEventRefresh, *errmsg.ErrMsg) {
	event := &dao.MapEventRefresh{RoleId: roleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), event)
	if err != nil {
		return event, err
	}
	return event, nil
}

func SaveEventRefresh(ctx *ctx.Context, event *dao.MapEventRefresh) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), event)
}

func getMapStoryKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.MapStory, values.Hash, roleId)
}

func SaveAppointMapEvent(ctx *ctx.Context, roleId values.RoleId, mapId int64, events []*models.AppointMapEvent) {
	data := &dao.AppointEvent{
		MapId:  mapId,
		Events: events,
	}
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), GetAppointMapEventKey(roleId), data)
}

func GetAppointMapEventKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.AppointMapEvent, values.Hash, roleId)
}

func LoadAppointMapEvent(ctx *ctx.Context, roleId values.RoleId, mapId int64) ([]*models.AppointMapEvent, *errmsg.ErrMsg) {
	ret := &dao.AppointEvent{
		MapId: mapId,
	}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), GetAppointMapEventKey(roleId), ret)
	if err != nil {
		return nil, err
	}
	if !has {
		return []*models.AppointMapEvent{}, nil
	}
	return ret.Events, nil
}
