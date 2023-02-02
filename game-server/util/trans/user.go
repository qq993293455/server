package trans

import (
	"math/rand"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/rule"
)

func TitleRewardsD2M(d *dao.TitleRewards) *models.TitleRewards {
	return &models.TitleRewards{
		Title:   d.Title,
		Rewards: d.Rewards,
	}
}

func Hero2CppHero(ctx *ctx.Context, h *models.Hero, equips map[values.EquipId]*models.Equipment) *models.HeroForBattle {
	hero := &models.HeroForBattle{}
	equip := make(map[int64]int64)
	equipLightEffect := make(map[int64]int64)
	for slot, item := range h.EquipSlot {
		if item.EquipItemId == 0 {
			equip[slot] = -1
			continue
		}
		equip[slot] = item.EquipItemId
		if equipModel, ok := equips[item.EquipId]; ok && equipModel.Detail != nil {
			if equipModel.Detail.LightEffect > 0 {
				equipLightEffect[slot] = equipModel.Detail.LightEffect
			}
		}
	}
	hero.SkillIds = h.Skill
	hero.Equip = equip
	hero.Attr = h.Attrs
	hero.ConfigId = h.Id
	hero.BuffIds = h.Buff
	hero.TalentBuff = h.TalentBuff
	hero.EquipLightEffect = equipLightEffect
	// hero.Fashion = h.Fashion.Dressed
	hero.Fashion = GetNewModelIdByFashionId(ctx, h.Fashion.Dressed)
	return hero
}

func Heroes2CppHeroes(ctx *ctx.Context, hs []*models.Hero, equips map[values.EquipId]*models.Equipment) []*models.HeroForBattle {
	heroes := make([]*models.HeroForBattle, 0, len(hs))
	for _, h := range hs {
		heroes = append(heroes, Hero2CppHero(ctx, h, equips))
	}
	return heroes
}

func GenPlayerName(uid int64) string {
	return "Player" + strconv.FormatInt(uid, 10)
}

func RandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func GetNewModelIdByFashionId(ctx *ctx.Context, fashionId values.FashionId) values.Integer {
	f, ok := rule.MustGetReader(ctx).Fashion.GetFashionById(fashionId)
	if ok {
		return f.NewModelId
	}
	return 0
}
