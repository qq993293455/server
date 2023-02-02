package tower

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/rule"
	"math/rand"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log *logger.Logger
}

func NewOptionalBoxService(
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
	module.OptionalBoxService = s
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("开启 宝箱", s.OpenBox)
}

//msg proc ============================================================================================================================================================================================================
func (this_ *Service) OpenBox(ctx *ctx.Context, req *servicepb.OptionalBox_OpenRequest) (*servicepb.OptionalBox_OpenResponse, *errmsg.ErrMsg) {
	cnf, ok := rule.MustGetReader(ctx).OptionalBox.GetOptionalBoxById(req.Item.ItemId)
	if !ok {
		return nil, errmsg.NewErrOptionalBoxExist()
	}

	if cnf.Typ != int64(req.Typ) {
		return nil, errmsg.NewErrOptionalBoxType()
	}

	if len(cnf.Item()) == 0 || len(cnf.Num()) != len(cnf.Item()) {
		return nil, errmsg.NewErrOptionalBoxCnf()
	}

	switch req.Typ {
	case models.OptionalBoxType_OptionalBox_Select:
		{
			if len(req.SelectItem) == 0 {
				return nil, errmsg.NewErrOptionalBoxNoSelect()
			}

			total := 0
			for _, item := range req.SelectItem {
				total += int(item.Count)

				hasItem := false
				for _, itemId := range cnf.Item() {
					if itemId == item.ItemId {
						hasItem = true
						break
					}
				}

				if !hasItem {
					this_.log.Error("Option box err", zap.Any("roleid", ctx.RoleId), zap.Any("select item", item.ItemId), zap.Any("req item", req.Item.ItemId))
					return nil, errmsg.NewErrOptionalBoxNoSelectItem()
				}
			}

			if total != int(req.Item.Count) {
				return nil, errmsg.NewErrOptionalBoxSelect()
			}
		}
	case models.OptionalBoxType_OptionalBox_Random:
		if len(cnf.Pro()) != len(cnf.Item()) {
			return nil, errmsg.NewErrOptionalBoxRandom()
		}
	}

	cnt, err := this_.BagService.GetItem(ctx, ctx.RoleId, req.Item.ItemId)
	if err != nil {
		return nil, err
	}

	if cnt < req.Item.Count {
		return nil, errmsg.NewErrOptionalBoxNotEnough()
	}

	ret := &servicepb.OptionalBox_OpenResponse{}

	this_.BagService.SubItem(ctx, ctx.RoleId, req.Item.ItemId, req.Item.Count)
	switch req.Typ {
	case models.OptionalBoxType_OptionalBox_Select:
		{
			for _, item := range req.SelectItem {
				for index, itemId := range cnf.Item() {
					if itemId == item.ItemId {
						this_.BagService.AddItem(ctx, ctx.RoleId, itemId, cnf.Num()[index]*item.Count)
						ret.Items = append(ret.Items, &models.Item{
							ItemId: itemId,
							Count:  cnf.Num()[index] * item.Count,
						})
						break
					}
				}
			}
		}
	case models.OptionalBoxType_OptionalBox_Random:
		{
			totalPro := int64(0)
			for _, pro := range cnf.Pro() {
				totalPro += pro
			}

			var items map[values.ItemId]values.Integer = make(map[int64]int64)

			for i := req.Item.Count; i > 0; i-- {
				rPro := rand.Int63n(totalPro)

				index := 0
				for pIndex, pro := range cnf.Pro() {
					rPro -= pro
					if rPro <= 0 {
						index = pIndex
						break
					}
				}

				items[cnf.Item()[index]] += cnf.Num()[index]
			}

			this_.BagService.AddManyItem(ctx, ctx.RoleId, items)

			for itemId, count := range items {
				ret.Items = append(ret.Items, &models.Item{
					ItemId: itemId,
					Count:  count,
				})
			}
		}
	}

	return ret, nil
}
