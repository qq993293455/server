package user

import (
	"errors"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/service/user/db"
	rule2 "coin-server/game-server/service/user/rule"
	"coin-server/rule"
)

func (this_ *Service) getOwnAvatar(c *ctx.Context) (*dao.RoleOwnAvatar, *errmsg.ErrMsg) {
	oa, err := db.GetOwnAvatar(c)
	if err != nil {
		return nil, err
	}
	for k, v := range oa.OwnAvatar {
		if v.ExpirationTime == 0 || v.ExpirationTime > timer.Now().UnixMilli() {
			continue
		}
		delete(oa.OwnAvatar, k)
	}
	return oa, nil
}

func (this_ *Service) AddAvatar(c *ctx.Context, itemId, avatarId, expirationTime values.Integer) *errmsg.ErrMsg {
	var needPush bool
	bm := rule.MustGetReader(c).GetBeginningMap()
	_, ok := bm[itemId]
	needPush = !ok

	oa, err := this_.getOwnAvatar(c)
	if err != nil {
		return err
	}

	avatar := &models.Avatar{Id: avatarId}
	if expirationTime > 0 {
		avatar.ExpirationTime = expirationTime * 1000
	} else if expirationTime < 0 {
		avatar.ExpirationTime = timer.StartTime(c.StartTime).UnixMilli() - expirationTime*1000
	}

	old := oa.OwnAvatar[avatarId]
	if old != nil && old.ExpirationTime == 0 { // 如果已经有永久的了
		return nil
	}
	if old == nil || old.ExpirationTime < avatar.ExpirationTime { // 如果未拥有 || 如果老的过期时间小于新的
		oa.OwnAvatar[avatarId] = avatar
		if needPush {
			c.PushMessage(&lessservicepb.User_GotAvatarPush{Avatar: avatar, ItemId: itemId})
		}
		db.SaveOwnAvatar(c, oa)
		return nil
	}

	return nil
}

//CheckAvatarExpired 检查头像头像框是否过期
func (this_ *Service) CheckAvatarExpired(c *ctx.Context) (*dao.RoleOwnAvatar, *models.CurAvatar, *errmsg.ErrMsg) {
	oa, err := this_.getOwnAvatar(c)
	if err != nil {
		return nil, nil, err
	}

	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, nil, err
	}
	if r == nil {
		return nil, nil, errmsg.NewErrUserNotFound()
	}

	var isChange bool
	originHead := rule2.MustGetInitialUseOfAvatar(c)
	if _, ok := oa.OwnAvatar[r.AvatarId]; !ok {
		r.AvatarId = originHead[0]
		isChange = true
	}
	if _, ok := oa.OwnAvatar[r.AvatarFrame]; !ok {
		r.AvatarFrame = originHead[1]
		isChange = true
	}
	if isChange {
		db.SaveRole(c, r)
	}
	return oa, &models.CurAvatar{
		IsChange:       isChange,
		CurAvatar:      r.AvatarId,
		CurAvatarFrame: r.AvatarFrame,
	}, nil
}

func (this_ *Service) GetOwnAvatar(c *ctx.Context, _ *lessservicepb.User_GetOwnAvatarRequest) (*lessservicepb.User_GetOwnAvatarResponse, *errmsg.ErrMsg) {
	oa, cur, err := this_.CheckAvatarExpired(c)
	if err != nil {
		return nil, err
	}
	return &lessservicepb.User_GetOwnAvatarResponse{
		OwnAvatar: oa.OwnAvatar,
		CurAvatar: cur,
	}, nil
}

func (this_ *Service) SetAvatar(c *ctx.Context, req *lessservicepb.User_SetAvatarRequest) (*lessservicepb.User_SetAvatarResponse, *errmsg.ErrMsg) {
	oa, err := this_.getOwnAvatar(c)
	if err != nil {
		return nil, err
	}
	if _, ok := oa.OwnAvatar[req.AvatarId]; !ok {
		return nil, errmsg.NewErrNotOwnedAvatar()
	}
	cfg, ok := rule.MustGetReader(c).HeadSculpture.GetHeadSculptureById(req.AvatarId)
	if !ok {
		panic(errors.New("HeadSculpture config not found: " + strconv.Itoa(int(req.AvatarId))))
	}
	if cfg.HeadSculptureType != 1 {
		return nil, errmsg.NewErrNotOwnedAvatar()
	}

	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}

	r.AvatarId = req.AvatarId
	db.SaveRole(c, r)

	return &lessservicepb.User_SetAvatarResponse{AvatarId: r.AvatarId}, nil
}

func (this_ *Service) SetAvatarFrame(c *ctx.Context, req *lessservicepb.User_SetAvatarFrameRequest) (*lessservicepb.User_SetAvatarFrameResponse, *errmsg.ErrMsg) {
	oa, err := this_.getOwnAvatar(c)
	if err != nil {
		return nil, err
	}
	if _, ok := oa.OwnAvatar[req.AvatarFrame]; !ok {
		return nil, errmsg.NewErrNotOwnedAvatarFrame()
	}
	cfg, ok := rule.MustGetReader(c).HeadSculpture.GetHeadSculptureById(req.AvatarFrame)
	if !ok {
		panic(errors.New("HeadSculpture config not found: " + strconv.Itoa(int(req.AvatarFrame))))
	}
	if cfg.HeadSculptureType != 2 {
		return nil, errmsg.NewErrNotOwnedAvatarFrame()
	}

	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	r.AvatarFrame = req.AvatarFrame
	db.SaveRole(c, r)

	return &lessservicepb.User_SetAvatarFrameResponse{AvatarFrame: r.AvatarFrame}, nil
}
