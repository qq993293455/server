package values

import (
	"math/rand"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/common/values"
)

type ManualShopI interface {
	Buy(idx, showTyp values.Integer) (*models.ShopDetail, *errmsg.ErrMsg)
	GetDetail() map[values.Integer]*models.TypeShopList
	Refresh(c *ctx.Context, lvl values.Level) *errmsg.ErrMsg
	CheckUpdate(ctx *ctx.Context, tbg values.Integer) bool
	AutoUpdate(ctx *ctx.Context, lvl values.Level, tbg values.Integer) *errmsg.ErrMsg
	TodayRefreshCnt() values.Integer
	LastRefreshAt() values.Integer
	ToDao() *dao.ManualUpdateShop

	CheatClearRefreshCnt(ctx *ctx.Context, tbg values.Integer)
	CheatRefresh(lvl values.Level) *errmsg.ErrMsg
}

type manualShop struct {
	typ    ShopTyp
	values *dao.ManualUpdateShop
}

func NewManualShop(values *dao.ManualUpdateShop, typ ShopTyp) ManualShopI {
	return &manualShop{values: values, typ: typ}
}

func (this_ *manualShop) Buy(idx, showTyp values.Integer) (*models.ShopDetail, *errmsg.ErrMsg) {
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

func (this_ *manualShop) CheckUpdate(ctx *ctx.Context, tbg values.Integer) bool {
	return this_.values.LastRefreshAt < tbg
}

func (this_ *manualShop) AutoUpdate(ctx *ctx.Context, lvl values.Level, tbg values.Integer) *errmsg.ErrMsg {
	detail, err := genDetail(lvl, this_.typ)
	if err != nil {
		return err
	}
	this_.values.Detail = detail
	this_.values.LastRefreshAt = tbg
	this_.values.TodayRefreshCnt = 0
	return nil
}

func (this_ *manualShop) GetDetail() map[values.Integer]*models.TypeShopList {
	return this_.values.GetDetail()
}

func (this_ *manualShop) Refresh(c *ctx.Context, lvl values.Level) *errmsg.ErrMsg {
	detail, err := genDetail(lvl, this_.typ)
	if err != nil {
		return err
	}
	this_.values.Detail = detail
	this_.values.LastManualRefreshAt = timer.StartTime(c.StartTime).Unix()
	this_.values.TodayRefreshCnt++
	return nil
}

func (this_ *manualShop) TodayRefreshCnt() values.Integer {
	return this_.values.TodayRefreshCnt
}

func (this_ *manualShop) LastRefreshAt() values.Integer {
	return this_.values.LastRefreshAt
}

func (this_ *manualShop) CheatClearRefreshCnt(ctx *ctx.Context, tbg values.Integer) {
	this_.values.LastRefreshAt = tbg
	this_.values.TodayRefreshCnt = 0
}

func (this_ *manualShop) CheatRefresh(lvl values.Level) *errmsg.ErrMsg {
	if detail, err := genDetail(lvl, this_.typ); err != nil {
		return err
	} else {
		this_.values.Detail = detail
	}
	return nil
}

func (this_ *manualShop) ToDao() *dao.ManualUpdateShop {
	return this_.values
}

func randWeight(weightList []values.Integer, cnt int) []int {
	wl := make([]values.Integer, len(weightList))
	copy(wl, weightList)
	if cnt >= len(wl) {
		res := make([]int, cnt)
		for idx := range res {
			res[idx] = idx
		}
		return res
	}
	total := int64(0)
	for i, w := range wl {
		if w < 0 {
			wl[i] = total
			continue
		}
		total += w
		wl[i] = total
	}
	res, idx := make([]int, 0, cnt), 0
	for idx < cnt {
		r := rand.Int63n(total + 1)
		for i, w := range wl {
			if r <= w {
				res = append(res, i)
				idx++
				curr := weightList[i]
				for j := i; j < len(wl); j++ {
					wl[j] -= curr
				}
				total -= curr
				if total < 0 {
					return res
				}
				break
			}
		}
	}
	return res
}

func randCnt(itemNum [2]values.Integer) values.Integer {
	cnt := int(itemNum[1] - itemNum[0] + 1)
	if cnt <= 0 {
		return 0
	}
	return values.Integer(rand.Intn(cnt)) + itemNum[0]
}
