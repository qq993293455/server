package values

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type AutoShopI interface {
	Buy(idx, showTyp values.Integer) (*models.ShopDetail, *errmsg.ErrMsg)
	GetDetail() map[values.Integer]*models.TypeShopList
	CheckUpdate(ctx *ctx.Context, tbg values.Integer) bool
	AutoUpdate(ctx *ctx.Context, lvl values.Level, tbg values.Integer) *errmsg.ErrMsg
	LastRefreshAt() values.Integer
	NextRefreshAt(ctx *ctx.Context) int64
	ToDao() *dao.AutoUpdateShop

	CheatRefresh(lvl values.Level) *errmsg.ErrMsg
}

type autoShop struct {
	typ    ShopTyp
	values *dao.AutoUpdateShop
}

func NewAutoShop(values *dao.AutoUpdateShop, typ ShopTyp) AutoShopI {
	return &autoShop{values: values, typ: typ}
}

func (this_ *autoShop) Buy(idx, showTyp values.Integer) (*models.ShopDetail, *errmsg.ErrMsg) {
	if _, exist := this_.values.GetDetail()[showTyp]; !exist {
		return nil, errmsg.NewErrNoItemInShop()
	}
	typList := this_.values.GetDetail()[showTyp]
	if idx < 0 || int(idx) >= len(typList.List) {
		return nil, errmsg.NewErrNoItemInShop()
	}
	item := typList.List[idx]
	if item.IsSale {
		return nil, errmsg.NewErrItemAlreadySale()
	}
	item.IsSale = true
	return item, nil
}

func (this_ *autoShop) CheckUpdate(ctx *ctx.Context, tbg values.Integer) bool {
	return this_.values.LastRefreshAt < tbg
}

func (this_ *autoShop) AutoUpdate(ctx *ctx.Context, lvl values.Level, tbg values.Integer) *errmsg.ErrMsg {
	detail, err := genDetail(lvl, this_.typ)
	if err != nil {
		return err
	}
	this_.values.Detail = detail
	this_.values.LastRefreshAt = tbg
	return nil
}

func (this_ *autoShop) NextRefreshAt(ctx *ctx.Context) int64 {
	return 0
}

func (this_ *autoShop) GetDetail() map[values.Integer]*models.TypeShopList {
	return this_.values.GetDetail()
}

func (this_ *autoShop) LastRefreshAt() values.Integer {
	return this_.values.LastRefreshAt
}

func (this_ *autoShop) CheatRefresh(lvl values.Level) *errmsg.ErrMsg {
	if detail, err := genDetail(lvl, this_.typ); err != nil {
		return err
	} else {
		this_.values.Detail = detail
	}
	return nil
}

func (this_ *autoShop) ToDao() *dao.AutoUpdateShop {
	return this_.values
}
