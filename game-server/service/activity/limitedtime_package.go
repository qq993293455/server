package activity

import (
	"math"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	modelspb "coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/timer"
	"coin-server/common/utils/generic/slices"
	"coin-server/common/values"
	"coin-server/game-server/service/activity/dao"
	"coin-server/game-server/util"
	"coin-server/rule"
	tasktarget "coin-server/rule/factory/task-target"
	rulemodel "coin-server/rule/rule-model"
)

func (svc *Service) getLimitedTimePackageLocks(c *ctx.Context) (*daopb.LimitedTimePackageLocks, *errmsg.ErrMsg) {
	lockInfo, err := dao.GetLimitedTimePackageLocks(c)
	if err != nil {
		return nil, err
	}

	nextReset := util.DefaultNextRefreshTime().UnixMilli()
	if lockInfo.ResetAt >= nextReset {
		return lockInfo, nil
	}

	reader := rule.MustGetReader(c)
	lockInfo.ResetAt = nextReset
	for id := range lockInfo.Locks {
		cfg, ok := reader.ActivityLimitedtimePackage.GetActivityLimitedtimePackageById(id)
		if !ok {
			return nil, errmsg.NewInternalErr("ActivityLimitedTimePackage config not found: " + strconv.Itoa(int(id)))
		}
		if cfg.TaskReset == 1 { // 每日重置的
			lockInfo.Locks[id] = 0
		}
	}
	dao.SaveLimitedTimePackageLocks(c, lockInfo)

	return lockInfo, nil
}

func (svc *Service) getLimitedTimePackages(c *ctx.Context) (*daopb.LimitedTimePackages, *errmsg.ErrMsg) {
	packs, err := dao.GetLimitedTimePackages(c)
	if err != nil {
		return nil, err
	}

	var needSave bool
	for id, pack := range packs.Packages {
		if pack.End > timer.Now().UnixMilli() {
			continue
		}
		if pack.RealEnd > timer.Now().UnixMilli() { // 保留一段时间用于支付回调容错
			packs.KeepPackages = append(packs.KeepPackages, pack)
		}
		delete(packs.Packages, id)
		needSave = true
	}

	if needSave {
		dao.SaveLimitedTimePackages(c, packs)
	}

	return packs, nil
}

func (svc *Service) checkUnlock(c *ctx.Context, typ models.TaskType, target, count values.Integer, isAccumulate bool) *errmsg.ErrMsg {
	lockInfo, err := svc.getLimitedTimePackageLocks(c)
	if err != nil {
		return err
	}

	var needSave bool
	var packs *daopb.LimitedTimePackages
	reader := rule.MustGetReader(c)
	for id, curCount := range lockInfo.Locks {
		cfg, ok := reader.ActivityLimitedtimePackage.GetActivityLimitedtimePackageById(id)
		if !ok {
			return errmsg.NewInternalErr("ActivityLimitedTimePackage config not found: " + strconv.Itoa(int(id)))
		}
		params := tasktarget.ParseParam(cfg.PackageConditions)
		if params == nil {
			return errmsg.NewInternalErr("ActivityLimitedTimePackage config PackageConditions failed: " + strconv.Itoa(int(id)))
		}
		// 类型不匹配的 || 目标不一致的 || 已解锁的 跳过
		if params.TaskType != typ || params.Target != target || curCount >= params.Count {
			continue
		}

		if isAccumulate {
			curCount = count
		} else {
			curCount += count
		}
		if curCount >= params.Count {
			if packs == nil {
				packs, err = svc.getLimitedTimePackages(c)
				if err != nil {
					return err
				}
			}
			role, err := svc.UserService.GetRoleByRoleId(c, c.RoleId)
			if err != nil {
				return nil
			}
			err = svc.unlock(c, packs, cfg, role.Recharge)
			if err != nil {
				return err
			}
		}
		lockInfo.Locks[id] = curCount
		needSave = true
	}
	if packs != nil {
		dao.SaveLimitedTimePackages(c, packs)
	}
	if needSave {
		dao.SaveLimitedTimePackageLocks(c, lockInfo)
	}

	return nil
}

func (svc *Service) unlock(c *ctx.Context, packs *daopb.LimitedTimePackages, cfg *rulemodel.ActivityLimitedtimePackage, recharge values.Integer) (err *errmsg.ErrMsg) {
	if _, ok := packs.Packages[cfg.Id]; ok {
		return nil
	}

	reader := rule.MustGetReader(c)
	now := timer.Now().UnixMilli()
	pkg := &modelspb.LimitedTimePackage{
		Id:                 cfg.Id,
		Begin:              now,
		End:                now + cfg.DurationTime*1000,
		RealEnd:            timer.Now().AddDate(0, 0, 1).UnixMilli() + cfg.DurationTime*1000, // 延迟1天删除
		MainCityIcon:       cfg.MainCityIcon,
		Banner:             cfg.BannerIcon,
		Name:               cfg.PacksName,
		DescribeLanguageId: cfg.ActivityDescribeLanguageId,
		Bags:               make([]*modelspb.GiftBag, 0, 3),
	}
	for _, bagCfg := range reader.ActivityLimitedtimePackagePay.ListByParentId(cfg.Id) {
		if len(bagCfg.RechargeInterval) < 2 {
			panic("ActivityLimitedTimePackagePay config RechargeInterval len < 2: " + strconv.Itoa(int(cfg.Id)))
		}
		if bagCfg.RechargeInterval[1] == -1 {
			bagCfg.RechargeInterval[1] = math.MaxInt64
		}
		if recharge >= bagCfg.RechargeInterval[0] && recharge <= bagCfg.RechargeInterval[1] {
			pkg.Bags = append(pkg.Bags, &modelspb.GiftBag{
				Id:        bagCfg.Id,
				Items:     bagCfg.BagProps,
				ChargeId:  bagCfg.ChargeId,
				IsSoldOut: false,
				PayValue:  bagCfg.PayValue,
			})
		}
	}
	packs.Packages[cfg.Id] = pkg
	c.PushMessage(&servicepb.Activity_LimitTimePackagePush{Package: pkg})

	return
}

func (svc *Service) GetLimitedTimePackages(c *ctx.Context, _ *servicepb.Activity_GetLimitTimePackagesRequest) (*servicepb.Activity_GetLimitTimePackagesResponse, *errmsg.ErrMsg) {
	packs, err := svc.getLimitedTimePackages(c)
	if err != nil {
		return nil, err
	}
	return &servicepb.Activity_GetLimitTimePackagesResponse{Packages: packs.Packages}, nil
}

func (svc *Service) buyLimitedTimePackage(c *ctx.Context, packs *daopb.LimitedTimePackages, pack *models.LimitedTimePackage, bag *models.GiftBag, isKeep bool) *errmsg.ErrMsg {
	if bag.IsSoldOut { // 已购买
		return errmsg.NewErrActivityGiftNotExist()
	}

	_, err := svc.BagService.AddManyItem(c, c.RoleId, bag.Items)
	if err != nil {
		return err
	}
	bag.IsSoldOut = true

	allSoldOut := true
	for _, b := range pack.Bags {
		if !b.IsSoldOut {
			allSoldOut = false
			break
		}
	}
	if allSoldOut {
		if isKeep {
			packs.KeepPackages = slices.Delete(packs.KeepPackages, func(tp *models.LimitedTimePackage) bool {
				return tp.Id == pack.Id
			})
		} else {
			delete(packs.Packages, pack.Id)
		}
		cfg, ok := rule.MustGetReader(c).ActivityLimitedtimePackage.GetActivityLimitedtimePackageById(pack.Id)
		if !ok {
			return errmsg.NewInternalErr("ActivityLimitedTimePackage config not found: " + strconv.Itoa(int(pack.Id)))
		}
		if cfg.TaskReset == 2 { // 卖完重置的
			lockInfo, err := svc.getLimitedTimePackageLocks(c)
			if err != nil {
				return err
			}
			lockInfo.Locks[pack.Id] = 0
			dao.SaveLimitedTimePackageLocks(c, lockInfo)
		}
	}
	dao.SaveLimitedTimePackages(c, packs)

	c.PushMessage(&servicepb.Activity_BuyLimitTimePackagePush{
		Id:    pack.Id,
		BagId: bag.Id,
		Items: bag.Items,
	})
	return nil
}

func (svc *Service) CheatBuyLimitedTimePackage(c *ctx.Context, req *servicepb.Activity_CheatBuyLimitTimePackagesRequest) (*servicepb.Activity_CheatBuyLimitTimePackagesResponse, *errmsg.ErrMsg) {
	packs, err := svc.getLimitedTimePackages(c)
	if err != nil {
		return nil, err
	}

	pack, ok := packs.Packages[req.Id]
	if !ok {
		return nil, errmsg.NewErrActivityGiftNotExist()
	}
	if req.BagIdx < 0 || int(req.BagIdx) >= len(pack.Bags) {
		return nil, errmsg.NewErrActivityGiftNotExist()
	}

	bag := pack.Bags[int(req.BagIdx)]

	err = svc.buyLimitedTimePackage(c, packs, pack, bag, false)
	if err != nil {
		return nil, err
	}

	return &servicepb.Activity_CheatBuyLimitTimePackagesResponse{}, nil
}

func (svc *Service) BuyLimitedTimePackage(c *ctx.Context, pcId values.Integer) *errmsg.ErrMsg {
	packs, err := svc.getLimitedTimePackages(c)
	if err != nil {
		return err
	}

	for _, pack := range packs.Packages {
		for _, bag := range pack.Bags {
			if bag.ChargeId != pcId {
				continue
			}
			err = svc.buyLimitedTimePackage(c, packs, pack, bag, false)
			if err != nil {
				return err
			}
			return nil
		}
	}

	for _, pack := range packs.KeepPackages {
		for _, bag := range pack.Bags {
			if bag.ChargeId != pcId {
				continue
			}
			err = svc.buyLimitedTimePackage(c, packs, pack, bag, true)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return errmsg.NewErrActivityGiftNotExist()
}
