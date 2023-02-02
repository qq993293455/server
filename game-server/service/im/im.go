package im

import (
	"fmt"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/im"
	"coin-server/common/logger"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/utils/generic/slices"
	"coin-server/common/utils/imutil"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/common/values/enum/Notice"
	"coin-server/game-server/module"
	"coin-server/game-server/service/im/dao"
	"coin-server/game-server/util/trans"
	"coin-server/rule"

	json "github.com/json-iterator/go"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log *logger.Logger
}

func NewImService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
		log:        log,
	}
	module.ImService = s
	// 监听红包开启道具使用
	s.RegisterUpdaterById(enum.RedPack, func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg {
		_, err := s.sendGift(ctx, itemId)
		return err
	})
	// msgs := []string{
	//	"恭喜获得Garfield公会成功击杀[暗黑地狱使者]！",
	//	"玩家“流浪小野猪”运气爆棚，锻造出[寡妇制造者]。",
	//	"临冬城外发现大量[稀有晶矿]，各位勇士们可以前往碰碰运气。",
	//	"玩家“Player996”成功击杀从元素位面进犯的[大元素使]。",
	//	"服务器将于GMT+8的06:00重启，届时更新V3.1版本",
	//	"玩家“林佳佳”成功进入异位面，目前异位面已有23211名勇者正在和恶魔战斗。",
	// }
	// if serverId == 15 {
	//	go func() {
	//		rand.Seed(time.Now().UnixNano())
	//		for {
	//			err := im.DefaultClient.SendMessage(context.Background(), &im.Message{
	//				Type:      im.MsgTypeBroadcast,
	//				RoleID:    "system",
	//				RoleName:  "system",
	//				Content:   msgs[rand.Intn(len(msgs))],
	//				ParseType: im.ParseTypeSys,
	//			})
	//			if err != nil {
	//				fmt.Println("发送系统消息失败", err)
	//			}
	//			time.Sleep(time.Duration(rand.Intn(6)+10) * time.Minute)
	//		}
	//	}()
	// }
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取 IM Token", svc.GetToken)
	svc.svc.RegisterFunc("获取红包详情", svc.GetGift)
	svc.svc.RegisterFunc("批量获取红包详情", svc.GetGiftByIds)
	svc.svc.RegisterFunc("领取红包", svc.DrawGift)
	svc.svc.RegisterFunc("分享", svc.Share)
	svc.svc.RegisterFunc("分享其他物品,计数,不会发送聊天", svc.ShareItem)
	svc.svc.RegisterFunc("作弊发送红包", svc.CheatSendGift)
	svc.svc.RegisterFunc("作弊发送跑马灯消息", svc.CheatSendMarquee)
}

func (svc *Service) GetRoleInfo(ctx *ctx.Context) (*daopb.Role, *errmsg.ErrMsg) {
	role, err := svc.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	return role, nil
}

// GetGuildId 获取我所在的公会id
func (svc *Service) GetGuildId(ctx *ctx.Context) (string, *errmsg.ErrMsg) {
	guildId, err := svc.GuildService.GetGuildIdByRole(ctx)
	if err != nil {
		return "", err
	}
	return guildId, nil
}

func (svc *Service) GetToken(ctx *ctx.Context, _ *servicepb.Im_GetTokenRequest) (*servicepb.Im_GetTokenResponse, *errmsg.ErrMsg) {
	role, err := svc.GetRoleInfo(ctx)
	if err != nil {
		return nil, err
	}

	rooms := []string{strconv.Itoa(int(role.Language))}
	// if role.Language == 0 {
	//	rooms = []string{"1"}
	// }
	//
	// guildId, err := svc.GetGuildId(ctx)
	// if err != nil {
	//	return nil, err
	// }
	// if guildId != "" {
	//	rooms = append(rooms, guildId)
	// }

	token, err2 := im.DefaultClient.GetToken(ctx, ctx.RoleId, role.Nickname, rooms, imutil.GetIMRoleInfoExtra(role))
	if err2 != nil {
		return nil, errmsg.NewInternalErr(err2.Error())
	}
	return &servicepb.Im_GetTokenResponse{Token: token}, nil
}

// GetGift 获取单个红包信息
func (svc *Service) GetGift(ctx *ctx.Context, req *servicepb.Im_GetGiftRequest) (*servicepb.Im_GetGiftResponse, *errmsg.ErrMsg) {
	gift, err := dao.GetGift(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if gift == nil {
		return nil, errmsg.NewErrGiftNotFound()
	}
	dc, err := dao.GetDrawCounter(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.Im_GetGiftResponse{Gift: trans.GiftD2M(gift), Count: dc.Count}, nil
}

// GetGiftByIds 批量获取红包信息
func (svc *Service) GetGiftByIds(ctx *ctx.Context, req *servicepb.Im_GetGiftByIdsRequest) (*servicepb.Im_GetGiftByIdsResponse, *errmsg.ErrMsg) {
	gifts := make([]*models.Gift, 0, len(req.Ids))
	for _, id := range req.Ids {
		gift, err := dao.GetGift(ctx, id)
		if err != nil {
			return nil, err
		}
		if gift == nil {
			return nil, errmsg.NewErrGiftNotFound()
		}
		gifts = append(gifts, trans.GiftD2M(gift))
	}
	return &servicepb.Im_GetGiftByIdsResponse{Gifts: gifts}, nil
}

// DrawGift 领取红包
func (svc *Service) DrawGift(ctx *ctx.Context, req *servicepb.Im_DrawGiftRequest) (*servicepb.Im_DrawGiftResponse, *errmsg.ErrMsg) {
	// TODO lock?
	gift, err := dao.GetGift(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if gift == nil {
		return nil, errmsg.NewErrGiftNotFound()
	}
	isDraw := false
	for _, v := range gift.Records {
		if v.RoleId == ctx.RoleId {
			isDraw = true
			break
		}
	}

	if !isDraw && int(gift.DrawCount) < len(gift.Items) { // 能抢
		// 每日领取次数限制
		dc, err := dao.GetDrawCounter(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
		if dc.Count >= GetGiftDayLimit() {
			return nil, errmsg.NewErrGiftDrawLimit()
		}

		role, err := svc.GetRoleInfo(ctx)
		if err != nil {
			return nil, err
		}

		item := gift.Items[gift.DrawCount]
		err = svc.AddItem(ctx, ctx.RoleId, item.ItemId, item.Count)
		if err != nil {
			return nil, err
		}

		gift.Records = append(gift.Records, &daopb.DrawRecord{
			RoleId:   ctx.RoleId,
			Item:     item,
			DrawTime: timer.UnixMilli(),
			Role:     role,
		})
		gift.DrawCount++
		dao.SaveGift(ctx, gift)

		dc.Count++
		dao.SaveDrawCounter(ctx, dc)

		return &servicepb.Im_DrawGiftResponse{Gift: trans.GiftD2M(gift), Item: trans.ItemD2M(item)}, nil
	}

	return &servicepb.Im_DrawGiftResponse{Gift: trans.GiftD2M(gift)}, nil
}

func (svc *Service) sendGift(ctx *ctx.Context, giftNo values.Integer) (string, *errmsg.ErrMsg) {
	role, err1 := svc.GetRoleInfo(ctx)
	if err1 != nil {
		return "", err1
	}

	gift := GenGift(xid.New().String(), ctx.RoleId, giftNo)
	if gift == nil {
		return "", errmsg.NewErrGiftNotFound()
	}

	dao.SaveGift(ctx, gift)

	gs, err := json.MarshalToString(trans.GiftD2M(gift))
	if err != nil {
		return "", errmsg.NewInternalErr(err.Error())
	}
	err = im.DefaultClient.SendMessage(ctx, &im.Message{
		Type:      im.MsgTypeBroadcast,
		RoleID:    ctx.RoleId,
		RoleName:  role.Nickname,
		Content:   gs,
		ParseType: im.ParseTypeRedPack,
		Extra:     imutil.GetIMRoleInfoExtra(role),
	})
	if err != nil {
		return "", errmsg.NewInternalErr(err.Error())
	}

	return gift.GiftId, nil
}

// Share 目前仅分享装备
func (svc *Service) Share(ctx *ctx.Context, req *servicepb.Im_ShareRequest) (*servicepb.Im_ShareResponse, *errmsg.ErrMsg) {
	if len(req.Id) <= 0 {
		return &servicepb.Im_ShareResponse{}, nil
	}
	role, err := svc.GetRoleInfo(ctx)
	if err != nil {
		return nil, err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, req.Id...)
	if err != nil {
		return nil, err
	}
	heroes, err := svc.GetAllHero(ctx, ctx.RoleId)
	equipList := make([]*Equipment, 0)
	starMap := make(map[values.EquipId]values.Integer)
	// equipList := make([]*models.Equipment, 0)
	for _, equipId := range req.Id {
		equip, ok := equips[equipId]
		if !ok {
			continue
		}
		starMap[equipId] = 0
		for _, hero := range heroes {
			if hero.Id == equip.HeroId {
				for _, slot := range hero.EquipSlot {
					if slot.EquipId == equipId {
						starMap[equipId] = slot.Star
						break
					}
				}
				break
			}
		}
		// e := (*Equipment)(unsafe.Pointer(equip))
		equipList = append(equipList, PB2Struct(equip))
	}
	if len(equipList) <= 0 {
		return &servicepb.Im_ShareResponse{}, nil
	}
	content, err1 := json.MarshalToString(equipList)
	if err1 != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	var star string
	if len(starMap) > 0 {
		star, err1 = json.MarshalToString(starMap)
		if err1 != nil {
			return nil, errmsg.NewInternalErr(err.Error())
		}
	}
	if err := im.DefaultClient.SendMessage(ctx, &im.Message{
		Type:      im.MsgTypeBroadcast,
		RoleID:    ctx.RoleId,
		RoleName:  role.Nickname,
		Content:   content,
		ParseType: im.ParseTypeShare,
		Extra:     imutil.GetShareEquipExtra(role, star),
	}); err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskShareEquip, 0, 1)
	return &servicepb.Im_ShareResponse{}, nil
}

// ShareItem 分享其他物品(只用作统计计数)，不会发送聊天
func (svc *Service) ShareItem(ctx *ctx.Context, _ *servicepb.Im_ShareItemRequest) (*servicepb.Im_ShareItemResponse, *errmsg.ErrMsg) {
	svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskShareEquip, 0, 1)
	return &servicepb.Im_ShareItemResponse{}, nil
}

// CheatSendGift 作弊器发送红包
func (svc *Service) CheatSendGift(ctx *ctx.Context, req *servicepb.Im_CheatSendGiftRequest) (*servicepb.Im_CheatSendGiftResponse, *errmsg.ErrMsg) {
	giftID, err := svc.sendGift(ctx, req.No)
	if err != nil {
		return nil, err
	}
	return &servicepb.Im_CheatSendGiftResponse{Id: giftID}, nil
}

// CheatSendMarquee 作弊器发送跑马灯消息
func (svc *Service) CheatSendMarquee(ctx *ctx.Context, req *servicepb.Im_CheatSendMarqueeRequest) (*servicepb.Im_CheatSendMarqueeResponse, *errmsg.ErrMsg) {
	err := im.DefaultClient.SendMessage(ctx, &im.Message{
		Type:       im.MsgTypeBroadcast,
		RoleID:     "admin",
		RoleName:   "admin",
		Content:    req.Message,
		ParseType:  im.ParseTypeSys,
		IsMarquee:  true,
		IsVolatile: req.IsVolatile,
	})
	if err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	return &servicepb.Im_CheatSendMarqueeResponse{}, nil
}

// SendNotice 发跑马灯公告
func (svc *Service) SendNotice(c *ctx.Context, parseType int, noticeId Notice.Enum, args ...any) *errmsg.ErrMsg {
	cfg, ok := rule.MustGetReader(nil).Notice.GetNoticeById(noticeId)
	if !ok {
		panic(fmt.Sprintf("notice config not found : %d", noticeId))
	}

	switch cfg.Typ {
	case 1: // 全服消息
		err := imutil.SendNotice(c, parseType, cfg.IsShowChat && slices.In(cfg.Channels, 1), noticeId, args...)
		if err != nil {
			return errmsg.NewInternalErr(err.Error())
		}
	case 2: // 战斗分线消息
		curRes, err1 := svc.BattleService.GetCurrBattleInfo(c, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
		if err1 != nil {
			return err1
		}
		line := imutil.BattleLineRoom(curRes.LineInfo.BattleServerId)
		err := imutil.SendNoticeByBattleLine(c, line, parseType, cfg.IsShowChat && slices.In(cfg.Channels, 1), noticeId, args...)
		if err != nil {
			return errmsg.NewInternalErr(err.Error())
		}
	}

	content := imutil.GenNoticeContent(noticeId, args...)
	for _, ch := range cfg.Channels {
		switch ch {
		case 2: // 公会
			guildId, err := svc.GuildService.GetGuildIdByRole(c)
			if err != nil {
				return err
			}
			imErr := im.DefaultClient.SendMessage(c, &im.Message{
				Type:      im.MsgTypeRoom,
				RoleID:    "admin",
				RoleName:  "admin",
				TargetID:  guildId,
				Content:   content,
				ParseType: parseType,
			})
			if imErr != nil {
				return errmsg.NewInternalErr(imErr.Error())
			}
		default:
			svc.log.Warn("not implemented notice to im channel", zap.Int64("channel", ch))
		}
	}
	return nil
}
