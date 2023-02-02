package friend

import (
	"sort"
	"strings"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	lesssvcpb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/module"
	"coin-server/game-server/service/friend/dao"
	"coin-server/rule"
)

const (
	FullFriendLimit        = "FriendCountLimit"       // 玩家好友上限,赠送上限
	DayRecvLimit           = "FriendPointsClaimLimit" // 每天收取上限
	DaySendLimit           = "FriendshipPointGiftCap" // 每日领取上限
	FriendPointsClaimPoint = "FriendPointsClaimPoint" // 每次领取好友点数的值

	AllRoleKey = "all_role_key" // 前端传全领
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	module     *module.Module
}

func NewFriendService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		module:     module,
	}
	module.FriendService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("获取好友列表", this_.GetFriendListRequest)
	this_.svc.RegisterFunc("获取申请列表", this_.GetRequestListRequest)
	this_.svc.RegisterFunc("获取黑名单列表", this_.GetBlackListRequest)
	this_.svc.RegisterFunc("发送加好友申请", this_.AddRequestRequest)
	this_.svc.RegisterFunc("确认好友申请", this_.ConfirmRequestRequest)
	this_.svc.RegisterFunc("删除好友", this_.DeleteRequest)
	this_.svc.RegisterFunc("加黑名单", this_.AddBlackRequest)
	this_.svc.RegisterFunc("移除黑名单", this_.RemoveBlackRequest)
	this_.svc.RegisterFunc("发送友情点", this_.SendPointRequest)
	this_.svc.RegisterFunc("接受友情点", this_.RecvPointRequest)
	this_.svc.RegisterFunc("赠送并收取友情点", this_.SendAndRecvPoint)

	this_.svc.RegisterEvent("服务器内部使用,发送友情点", this_.ServerSendPointPush)

	this_.svc.RegisterFunc("作弊器清除发送和收到的友情点", this_.CheatClearSendAndRecv)
}

//---------------------------------------------------module------------------------------------------------------------//

func (this_ *Service) GetFriendIds(ctx *ctx.Context) ([]values.RoleId, *errmsg.ErrMsg) {
	friendData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	roleIds := make([]values.RoleId, 0, len(friendData.Friends))
	for _, f := range friendData.Friends {
		roleIds = append(roleIds, f.RoleId)
	}
	return roleIds, nil
}

//---------------------------------------------------proto------------------------------------------------------------//

func (this_ *Service) GetFriendListRequest(ctx *ctx.Context, _ *lesssvcpb.Friend_GetFriendListRequest) (*lesssvcpb.Friend_GetFriendListResponse, *errmsg.ErrMsg) {
	friendData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	list, roleIds, idx := make([]*models.FriendInfo, len(friendData.Friends)), make([]values.RoleId, len(friendData.Friends)), 0
	beginOfDay := this_.module.RefreshService.GetCurrDayFreshTime(ctx).Unix()
	isTodaySend, isTodayRecv := friendData.LastSendAt >= beginOfDay, friendData.LastRecvAt >= beginOfDay
	isChange := false
	if !isTodaySend {
		friendData.TodaySend = 0
		friendData.LastSendAt = timer.Unix()
		isChange = true
	}
	if !isTodayRecv {
		friendData.TodayRecv = 0
		friendData.LastRecvAt = timer.Unix()
		isChange = true
	}
	for _, v := range friendData.Friends {
		if !isTodaySend {
			v.IsSend = false
		}
		if !isTodayRecv {
			if v.IsRecv == pbdao.RecvGiftType_get {
				v.IsRecv = pbdao.RecvGiftType_neither
			}
		}
		list[idx] = &models.FriendInfo{
			RoleId:   v.RoleId,
			CreateAt: v.CreateAt,
			IsRecv:   models.RecvGiftType(v.IsRecv),
			IsSend:   v.IsSend,
		}
		roleIds[idx] = v.RoleId
		idx++
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreateAt < list[j].CreateAt
	})
	roleMap, err := this_.module.GetRole(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if r, exist := roleMap[v.RoleId]; exist {
			v.NickName = r.Nickname
			v.Power = r.Power
			v.Avatar = r.AvatarId
			v.AvatarFrame = r.AvatarFrame
			v.Lvl = r.Level
			v.LastLoginAt = r.Login
			v.LastLogoutAt = r.Logout
		}
	}
	if isChange {
		dao.Save(ctx, friendData)
	}
	return &lesssvcpb.Friend_GetFriendListResponse{
		FriendList: list,
		TodayRecv:  friendData.TodayRecv,
		TodaySend:  friendData.TodaySend,
	}, nil
}

func (this_ *Service) GetRequestListRequest(ctx *ctx.Context, _ *lesssvcpb.Friend_GetRequestListRequest) (*lesssvcpb.Friend_GetRequestListResponse, *errmsg.ErrMsg) {
	friendData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	list, roleIds, idx := make([]*models.RequestInfo, len(friendData.Requests)), make([]values.RoleId, len(friendData.Requests)), 0
	for _, v := range friendData.Requests {
		list[idx] = &models.RequestInfo{
			RoleId:   v.RoleId,
			CreateAt: v.CreateAt,
		}
		roleIds[idx] = v.RoleId
		idx++
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreateAt < list[j].CreateAt
	})
	roleMap, err := this_.module.GetRole(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if r, exist := roleMap[v.RoleId]; exist {
			v.NickName = r.Nickname
			v.Power = r.Power
			v.Avatar = r.AvatarId
			v.AvatarFrame = r.AvatarFrame
			v.Lvl = r.Level
		}
	}
	return &lesssvcpb.Friend_GetRequestListResponse{
		RequestList: list,
	}, nil
}

func (this_ *Service) GetBlackListRequest(ctx *ctx.Context, _ *lesssvcpb.User_GetBlackListRequest) (*lesssvcpb.User_GetBlackListResponse, *errmsg.ErrMsg) {
	friendData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	list, roleIds, idx := make([]*models.BlackListInfo, len(friendData.Blacklist)), make([]values.RoleId, len(friendData.Blacklist)), 0
	for _, v := range friendData.Blacklist {
		list[idx] = &models.BlackListInfo{
			RoleId:   v.RoleId,
			CreateAt: v.CreateAt,
		}
		roleIds[idx] = v.RoleId
		idx++
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreateAt < list[j].CreateAt
	})
	roleMap, err := this_.module.GetRole(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if r, exist := roleMap[v.RoleId]; exist {
			v.NickName = r.Nickname
			v.Power = r.Power
			v.Avatar = r.AvatarId
			v.AvatarFrame = r.AvatarFrame
			v.Lvl = r.Level
		}
	}
	return &lesssvcpb.User_GetBlackListResponse{
		Blacklist: list,
	}, nil
}

func (this_ *Service) AddRequestRequest(ctx *ctx.Context, req *lesssvcpb.Friend_AddRequestRequest) (*lesssvcpb.Friend_AddRequestResponse, *errmsg.ErrMsg) {
	if req.RoleId == "" || req.RoleId == ctx.RoleId {
		return nil, errmsg.NewErrTarIllegal()
	}
	err := ctx.DRLock(redisclient.GetLocker(), getLockKey(req.RoleId), getLockKey(ctx.RoleId))
	if err != nil {
		return nil, err
	}
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if _, exist := myData.Friends[req.RoleId]; exist {
		return nil, errmsg.NewErrFriendExist()
	}
	limit, ok := rule.MustGetReader(ctx).KeyValue.GetInt(FullFriendLimit)
	if !ok {
		limit = 0
	}
	if len(myData.Friends) >= limit {
		return nil, errmsg.NewErrFriendFull()
	}
	tarData, err := dao.Get(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if len(tarData.Friends) >= limit {
		return nil, errmsg.NewErrTarFriendFull()
	}
	if _, exist := tarData.Blacklist[ctx.RoleId]; exist {
		return nil, errmsg.NewErrBlackExist()
	}
	if _, exist := tarData.Requests[ctx.RoleId]; exist {
		return nil, errmsg.NewErrRequestExist()
	}
	if _, exist := myData.Blacklist[req.RoleId]; exist {
		delete(myData.Blacklist, req.RoleId)
	}
	tarData.Requests[ctx.RoleId] = &pbdao.RequestValue{
		RoleId:   ctx.RoleId,
		CreateAt: timer.Now().UnixMilli(),
	}
	dao.Save(ctx, myData)
	dao.Save(ctx, tarData)
	ctx.PushMessageToRole(req.RoleId, &lesssvcpb.Friend_RequestListUpdatePush{})
	return &lesssvcpb.Friend_AddRequestResponse{}, nil
}

func (this_ *Service) AddBlackRequest(ctx *ctx.Context, req *lesssvcpb.Friend_AddBlackRequest) (*lesssvcpb.Friend_AddBlackResponse, *errmsg.ErrMsg) {
	if req.RoleId == "" || req.RoleId == ctx.RoleId {
		return nil, errmsg.NewErrTarIllegal()
	}
	err := ctx.DRLock(redisclient.GetLocker(), getLockKey(req.RoleId), getLockKey(ctx.RoleId))
	if err != nil {
		return nil, err
	}
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if _, exist := myData.Blacklist[req.RoleId]; exist {
		return nil, errmsg.NewErrBlackExist()
	}
	tarData, err := dao.Get(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if _, exist := myData.Friends[req.RoleId]; exist {
		delete(myData.Friends, req.RoleId)
	}
	if _, exist := tarData.Friends[ctx.RoleId]; exist {
		delete(tarData.Friends, ctx.RoleId)
		ctx.PushMessageToRole(req.RoleId, &lesssvcpb.Friend_FriendListUpdatePush{})
	}
	myData.Blacklist[req.RoleId] = &pbdao.BlackListValue{
		RoleId:   req.RoleId,
		CreateAt: timer.Now().UnixMilli(),
	}
	dao.Save(ctx, myData)
	dao.Save(ctx, tarData)
	return &lesssvcpb.Friend_AddBlackResponse{}, nil
}

func (this_ *Service) ConfirmRequestRequest(ctx *ctx.Context, req *lesssvcpb.Friend_ConfirmRequestRequest) (*lesssvcpb.Friend_ConfirmRequestResponse, *errmsg.ErrMsg) {
	if req.RoleId == "" || req.RoleId == ctx.RoleId {
		return nil, errmsg.NewErrTarIllegal()
	}
	err := ctx.DRLock(redisclient.GetLocker(), getLockKey(req.RoleId), getLockKey(ctx.RoleId))
	if err != nil {
		return nil, err
	}
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if _, exist := myData.Requests[req.RoleId]; !exist {
		return nil, errmsg.NewErrRequestNotExist()
	}
	if !req.IsConform {
		delete(myData.Requests, req.RoleId)
		dao.Save(ctx, myData)
		return &lesssvcpb.Friend_ConfirmRequestResponse{}, nil
	}
	if _, exist := myData.Friends[req.RoleId]; exist {
		delete(myData.Requests, req.RoleId)
		dao.Save(ctx, myData)
		return &lesssvcpb.Friend_ConfirmRequestResponse{ErrCode: errmsg.NewErrFriendExist().ErrMsg}, nil
	}
	tarData, err := dao.Get(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if _, exist := tarData.Friends[ctx.RoleId]; exist {
		delete(myData.Requests, req.RoleId)
		dao.Save(ctx, myData)
		return &lesssvcpb.Friend_ConfirmRequestResponse{ErrCode: errmsg.NewErrFriendExist().ErrMsg}, nil
	}
	if _, exist := tarData.Blacklist[ctx.RoleId]; exist {
		delete(myData.Requests, req.RoleId)
		dao.Save(ctx, myData)
		return &lesssvcpb.Friend_ConfirmRequestResponse{ErrCode: errmsg.NewErrBlackExist().ErrMsg}, nil
	}
	if _, exist := myData.Blacklist[req.RoleId]; exist {
		delete(myData.Requests, req.RoleId)
		dao.Save(ctx, myData)
		return &lesssvcpb.Friend_ConfirmRequestResponse{ErrCode: errmsg.NewErrBlackExist().ErrMsg}, nil
	}
	limit, ok := rule.MustGetReader(ctx).KeyValue.GetInt(FullFriendLimit)
	if !ok {
		limit = 0
	}
	if len(myData.Friends) >= limit {
		return nil, errmsg.NewErrFriendFull()
	}
	if len(tarData.Friends) >= limit {
		return nil, errmsg.NewErrTarFriendFull()
	}
	delete(myData.Requests, req.RoleId)
	myData.Friends[req.RoleId] = &pbdao.FriendValue{
		RoleId:   req.RoleId,
		CreateAt: timer.Now().UnixMilli(),
	}
	tarData.Friends[ctx.RoleId] = &pbdao.FriendValue{
		RoleId:   ctx.RoleId,
		CreateAt: timer.Now().UnixMilli(),
	}
	dao.Save(ctx, myData)
	dao.Save(ctx, tarData)
	ctx.PushMessageToRole(req.RoleId, &lesssvcpb.Friend_FriendListUpdatePush{})
	this_.module.TaskService.UpdateTarget(ctx, req.RoleId, models.TaskType_TaskAddFriendCnt, 0, 1)
	this_.module.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskAddFriendCnt, 0, 1)
	return &lesssvcpb.Friend_ConfirmRequestResponse{}, nil
}

func (this_ *Service) DeleteRequest(ctx *ctx.Context, req *lesssvcpb.Friend_DeleteRequest) (*lesssvcpb.Friend_DeleteResponse, *errmsg.ErrMsg) {
	if req.RoleId == "" || req.RoleId == ctx.RoleId {
		return nil, errmsg.NewErrTarIllegal()
	}
	err := ctx.DRLock(redisclient.GetLocker(), getLockKey(req.RoleId), getLockKey(ctx.RoleId))
	if err != nil {
		return nil, err
	}
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if _, exist := myData.Friends[req.RoleId]; !exist {
		return nil, errmsg.NewErrFriendNotExist()
	}
	tarData, err := dao.Get(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if _, exist := tarData.Friends[ctx.RoleId]; !exist {
		return nil, errmsg.NewErrFriendNotExist()
	}
	delete(myData.Friends, req.RoleId)
	delete(tarData.Friends, ctx.RoleId)
	dao.Save(ctx, myData)
	dao.Save(ctx, tarData)
	ctx.PushMessageToRole(req.RoleId, &lesssvcpb.Friend_FriendListUpdatePush{})
	return &lesssvcpb.Friend_DeleteResponse{}, nil
}

func (this_ *Service) RemoveBlackRequest(ctx *ctx.Context, req *lesssvcpb.Friend_RemoveBlackRequest) (*lesssvcpb.Friend_RemoveBlackResponse, *errmsg.ErrMsg) {
	if req.RoleId == "" || req.RoleId == ctx.RoleId {
		return nil, errmsg.NewErrTarIllegal()
	}
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if _, exist := myData.Blacklist[req.RoleId]; !exist {
		return nil, errmsg.NewErrBlackNotExist()
	}
	delete(myData.Blacklist, req.RoleId)
	dao.Save(ctx, myData)
	return &lesssvcpb.Friend_RemoveBlackResponse{}, nil
}

func (this_ *Service) SendPointRequest(ctx *ctx.Context, req *lesssvcpb.Friend_SendPointRequest) (*lesssvcpb.Friend_SendPointResponse, *errmsg.ErrMsg) {
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	var roleIds []values.RoleId
	if req.RoleIds == AllRoleKey {
		roleIds = make([]values.RoleId, 0, len(myData.Friends))
		for _, friend := range myData.Friends {
			if !friend.IsSend {
				roleIds = append(roleIds, friend.RoleId)
			}
		}
	} else {
		roleIds = strings.Split(req.RoleIds, ",")
	}
	todayBegin := this_.module.RefreshService.GetCurrDayFreshTime(ctx).Unix()
	if myData.LastSendAt < todayBegin {
		myData.TodaySend = 0
		myData.LastSendAt = time.Unix(0, ctx.StartTime).UTC().Unix()
	}
	sendLimit, ok := rule.MustGetReader(ctx).KeyValue.GetInt64(DaySendLimit)
	if !ok {
		sendLimit = 0
	}
	userMap, err := this_.module.GetUserByRoleIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	isSend := false
	for _, tar := range roleIds {
		if myData.TodaySend >= sendLimit {
			break
		}
		if v, exist := myData.Friends[tar]; exist {
			if u, has := userMap[tar]; has {
				if !v.IsSend {
					isSend = true
					myData.Friends[tar].IsSend = true
					myData.TodaySend++
					ctx.PublishEventRemote(tar, u.ServerId, u.UserId, &lesssvcpb.Friend_ServerSendPointPush{
						SendRole: ctx.RoleId,
					})
				}
			}
		}
	}
	dao.Save(ctx, myData)
	ctx.PushMessage(&lesssvcpb.Friend_FriendListUpdatePush{})
	return &lesssvcpb.Friend_SendPointResponse{
		IsSend: isSend,
	}, nil
}

func (this_ *Service) RecvPointRequest(ctx *ctx.Context, req *lesssvcpb.Friend_RecvPointRequest) (*lesssvcpb.Friend_RecvPointResponse, *errmsg.ErrMsg) {
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	var roleIds []values.RoleId
	if req.RoleIds == AllRoleKey {
		roleIds = make([]values.RoleId, 0, len(myData.Friends))
		for _, friend := range myData.Friends {
			if friend.IsRecv == pbdao.RecvGiftType_recv_no_get {
				roleIds = append(roleIds, friend.RoleId)
			}
		}
	} else {
		roleIds = strings.Split(req.RoleIds, ",")
	}
	todayBegin := this_.module.RefreshService.GetCurrDayFreshTime(ctx).Unix()
	if myData.LastRecvAt < todayBegin {
		myData.TodayRecv = 0
		myData.LastRecvAt = time.Unix(0, ctx.StartTime).UTC().Unix()
	}
	recvLimit, ok := rule.MustGetReader(ctx).KeyValue.GetInt64(DayRecvLimit)
	if !ok {
		recvLimit = 0
	}
	point := int64(0)
	for _, tar := range roleIds {
		if myData.TodayRecv >= recvLimit {
			break
		}
		if _, exist := myData.Friends[tar]; exist {
			if myData.Friends[tar].IsRecv == pbdao.RecvGiftType_recv_no_get {
				myData.Friends[tar].IsRecv = pbdao.RecvGiftType_get
				myData.TodayRecv++
				point++
			}
		}
	}
	if point > 0 {
		if err = this_.module.AddItem(ctx, ctx.RoleId, enum.FriendPoint, point); err != nil {
			return nil, err
		}
	}
	dao.Save(ctx, myData)
	ctx.PushMessage(&lesssvcpb.Friend_FriendListUpdatePush{})
	return &lesssvcpb.Friend_RecvPointResponse{
		IsRecv: point > 0,
	}, nil
}

func (this_ *Service) SendAndRecvPoint(ctx *ctx.Context, _ *lesssvcpb.Friend_SendAndRecvRequest) (*lesssvcpb.Friend_SendAndRecvResponse, *errmsg.ErrMsg) {
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	roleIds := make([]values.RoleId, 0, len(myData.Friends))
	for _, friend := range myData.Friends {
		if !friend.IsSend || friend.IsRecv == pbdao.RecvGiftType_recv_no_get {
			roleIds = append(roleIds, friend.RoleId)
		}
	}
	todayBegin := this_.module.RefreshService.GetCurrDayFreshTime(ctx).Unix()
	if myData.LastSendAt < todayBegin {
		myData.TodaySend = 0
		myData.LastSendAt = time.Unix(0, ctx.StartTime).UTC().Unix()
	}
	userMap, err := this_.module.GetUserByRoleIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	sendLimit, ok := rule.MustGetReader(ctx).KeyValue.GetInt64(DaySendLimit)
	if !ok {
		sendLimit = 0
	}
	isSend := false
	for _, tar := range roleIds {
		if myData.TodaySend >= sendLimit {
			break
		}
		if v, exist := myData.Friends[tar]; exist {
			if u, has := userMap[tar]; has {
				if !v.IsSend {
					isSend = true
					myData.Friends[tar].IsSend = true
					myData.TodaySend++
					ctx.PublishEventRemote(tar, u.ServerId, u.UserId, &lesssvcpb.Friend_ServerSendPointPush{
						SendRole: ctx.RoleId,
					})
				}
			}
		}
	}
	if myData.LastRecvAt < todayBegin {
		myData.TodayRecv = 0
		myData.LastRecvAt = time.Unix(0, ctx.StartTime).UTC().Unix()
	}
	recvLimit, ok := rule.MustGetReader(ctx).KeyValue.GetInt64(DayRecvLimit)
	if !ok {
		recvLimit = 0
	}
	point := int64(0)
	for _, tar := range roleIds {
		if myData.TodayRecv >= recvLimit {
			break
		}
		if _, exist := myData.Friends[tar]; exist {
			if myData.Friends[tar].IsRecv == pbdao.RecvGiftType_recv_no_get {
				myData.Friends[tar].IsRecv = pbdao.RecvGiftType_get
				myData.TodayRecv++
				point++
			}
		}
	}
	if point > 0 {
		if err = this_.module.AddItem(ctx, ctx.RoleId, enum.FriendPoint, point); err != nil {
			return nil, err
		}
	}
	dao.Save(ctx, myData)
	ctx.PushMessage(&lesssvcpb.Friend_FriendListUpdatePush{})
	return &lesssvcpb.Friend_SendAndRecvResponse{
		IsRecv: point > 0,
		IsSend: isSend,
	}, nil
}

func (this_ *Service) CheatClearSendAndRecv(ctx *ctx.Context, req *lesssvcpb.Friend_CheatClearSendAndRecvRequest) (*lesssvcpb.Friend_CheatClearSendAndRecvResponse, *errmsg.ErrMsg) {
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	for _, friend := range myData.Friends {
		friend.IsRecv = pbdao.RecvGiftType_neither
		friend.IsSend = false
	}
	ctx.PushMessageToRole(ctx.RoleId, &lesssvcpb.Friend_FriendListUpdatePush{})
	myData.TodayRecv = 0
	myData.TodaySend = 0
	dao.Save(ctx, myData)
	return &lesssvcpb.Friend_CheatClearSendAndRecvResponse{}, nil
}

func (this_ *Service) ServerSendPointPush(ctx *ctx.Context, msg *lesssvcpb.Friend_ServerSendPointPush) {
	myData, err := dao.Get(ctx, ctx.RoleId)
	if err != nil {
		return
	}
	if tar, exist := myData.Friends[msg.SendRole]; exist {
		if tar.IsRecv == pbdao.RecvGiftType_neither {
			myData.Friends[msg.SendRole].IsRecv = pbdao.RecvGiftType_recv_no_get
		}
	}
	ctx.PushMessageToRole(ctx.RoleId, &lesssvcpb.Friend_FriendListUpdatePush{})
	dao.Save(ctx, myData)
}

//---------------------------------------------------util------------------------------------------------------------//

func getLockKey(roleId string) string {
	return "stateless:friend:" + roleId
}

func getLockKeys(roleIds []string) []string {
	res := make([]string, len(roleIds))
	for idx := range roleIds {
		res[idx] = "stateless:friend:" + roleIds[idx]
	}
	return res
}
