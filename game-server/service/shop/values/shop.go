package values

import (
	"coin-server/common/proto/dao"
	"coin-server/common/values"
)

type ShopTyp int64

const (
	ShopTypNormal ShopTyp = iota // 普通商店
	ShopTypArena                 // 竞技场商店
	ShopTypGuild                 // 公会商店
	ShopTypCamp                  // 势力商店
)

type Shop struct {
	RoleId values.RoleId
	Data   ManualShopI
}

func NewShop(data *dao.Shop) *Shop {
	return &Shop{
		RoleId: data.RoleId,
		Data:   NewManualShop(data.Data, ShopTypNormal),
	}
}

func (s *Shop) ToDao() *dao.Shop {
	return &dao.Shop{
		RoleId: s.RoleId,
		Data:   s.Data.ToDao(),
	}
}

type ArenaShop struct {
	RoleId values.RoleId
	Data   AutoShopI
}

func NewArenaShop(data *dao.ArenaShop) *ArenaShop {
	return &ArenaShop{
		RoleId: data.RoleId,
		Data:   NewAutoShop(data.Data, ShopTypArena),
	}
}

func (s *ArenaShop) ToDao() *dao.ArenaShop {
	return &dao.ArenaShop{
		RoleId: s.RoleId,
		Data:   s.Data.ToDao(),
	}
}

type GuildShop struct {
	RoleId values.RoleId
	Data   AutoShopI
}

func NewGuildShop(data *dao.GuildShop) *GuildShop {
	return &GuildShop{
		RoleId: data.RoleId,
		Data:   NewAutoShop(data.Data, ShopTypGuild),
	}
}

func (s *GuildShop) ToDao() *dao.GuildShop {
	return &dao.GuildShop{
		RoleId: s.RoleId,
		Data:   s.Data.ToDao(),
	}
}

type CampShop struct {
	RoleId values.RoleId
	Data   ManualShopI
}

func NewCampShop(data *dao.CampShop) *CampShop {
	return &CampShop{
		RoleId: data.RoleId,
		Data:   NewManualShop(data.Data, ShopTypCamp),
	}
}

func (s *CampShop) ToDao() *dao.CampShop {
	return &dao.CampShop{
		RoleId: s.RoleId,
		Data:   s.Data.ToDao(),
	}
}
