package im

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type Equipment struct {
	// 装备ID
	EquipId values.EquipId `json:"equipId"`
	// 装备模板ID
	ItemId    int64 `json:"itemId"`
	BaseScore int64
	// 物等
	Level  int64            `json:"level"`
	HeroId int64            `json:"heroId"`
	Detail *EquipmentDetail `json:"detail"`
}

type EquipmentDetail struct {
	Score int64 `json:"score"`
	// 词缀效果
	Affix       []*Affix `json:"affix"`
	ForgeId     string   `json:"forgeId"`
	ForgeName   string   `json:"forgeName"`
	LightEffect int64    `json:"lightEffect"`
}

type Affix struct {
	AffixId    int64           `json:"affixId"`
	Quality    int64           `json:"quality"`
	AffixValue int64           `json:"affixValue"`
	BuffId     int64           `json:"buffId"`
	Active     bool            `json:"active"`
	AttrId     int64           `json:"attrId"`
	Bonus      map[int64]int64 `json:"bonus"`
	IsPercent  bool            `json:"isPercent"`
}

func PB2Struct(equip *models.Equipment) *Equipment {
	return &Equipment{
		EquipId: equip.EquipId,
		ItemId:  equip.ItemId,
		// BaseScore: equip.BaseScore,
		Level:  equip.Level,
		HeroId: equip.HeroId,
		Detail: &EquipmentDetail{
			Score: equip.Detail.Score,
			Affix: func() []*Affix {
				list := make([]*Affix, 0)
				for _, affix := range equip.Detail.Affix {
					list = append(list, &Affix{
						AffixId:    affix.AffixId,
						Quality:    affix.Quality,
						AffixValue: affix.AffixValue,
						BuffId:     affix.BuffId,
						Active:     affix.Active,
						AttrId:     affix.AttrId,
						Bonus:      affix.Bonus,
						IsPercent:  affix.IsPercent,
					})
				}
				return list
			}(),
			// ForgeId:     equip.Detail.ForgeId,
			ForgeName:   equip.Detail.ForgeName,
			LightEffect: equip.Detail.LightEffect,
		},
	}
}
