package bag

import (
	"unsafe"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

const (
	DefaultRelicsLevel = 0
	DefaultRelicsStar  = 0
)

func ItemModel2Dao(itemModel *models.Item) *dao.Item {
	return (*dao.Item)(itemModel)
}

func ItemDao2Model(itemDao *dao.Item) *models.Item {
	return (*models.Item)(itemDao)
}

func ItemModels2Dao(itemModel []*models.Item) []*dao.Item {
	return *(*[]*dao.Item)(unsafe.Pointer(&itemModel))
}

func ItemDao2Models(itemDao []*dao.Item) []*models.Item {
	return *(*[]*models.Item)(unsafe.Pointer(&itemDao))
}

func EquipModel2Dao(equipModel *models.Equipment) *dao.Equipment {
	return &dao.Equipment{
		EquipId: equipModel.EquipId,
		ItemId:  equipModel.ItemId,
		Level:   equipModel.Level,
		Affix:   equipModel.Detail.Affix,
		HeroId:  equipModel.HeroId,
		// ForgeId:     equipModel.Detail.ForgeId,
		ForgeName:   equipModel.Detail.ForgeName,
		LightEffect: equipModel.Detail.LightEffect,
		// BaseScore:   equipModel.BaseScore,
		// Score:       equipModel.Detail.Score,
	}
}

func EquipDao2Model(equipDao *dao.Equipment, withDetail bool) *models.Equipment {
	equip := &models.Equipment{
		EquipId: equipDao.EquipId,
		ItemId:  equipDao.ItemId,
		// BaseScore: equipDao.BaseScore,
		Level:  equipDao.Level,
		HeroId: equipDao.HeroId,
	}
	if withDetail {
		equip.Detail = &models.EquipmentDetail{
			// Score:       equipDao.Score,
			Affix: equipDao.Affix,
			// ForgeId:     equipDao.ForgeId,
			ForgeName:   equipDao.ForgeName,
			LightEffect: equipDao.LightEffect,
		}
	}
	return equip
}

func EquipModels2Dao(equipModel []*models.Equipment) []*dao.Equipment {
	list := make([]*dao.Equipment, 0, len(equipModel))
	for _, equipment := range equipModel {
		list = append(list, EquipModel2Dao(equipment))
	}
	return list
}

// func EquipBriefDao2Models(equipDao []*dao.EquipmentBrief) []*models.Equipment {
// 	list := make([]*models.Equipment, 0, len(equipDao))
// 	for _, equipment := range equipDao {
// 		list = append(list, &models.Equipment{
// 			EquipId:   equipment.EquipId,
// 			ItemId:    equipment.ItemId,
// 			BaseScore: equipment.BaseScore,
// 			Level:     equipment.Level,
// 			HeroId:    equipment.HeroId,
// 		})
// 	}
// 	return list
// }

// func EquipModel2EquipmentBrief(equip *models.Equipment) *dao.EquipmentBrief {
// 	return &dao.EquipmentBrief{
// 		EquipId:   equip.EquipId,
// 		ItemId:    equip.ItemId,
// 		BaseScore: equip.BaseScore,
// 		Level:     equip.Level,
// 		HeroId:    equip.HeroId,
// 	}
// }

// func EquipManyModel2EquipmentBrief(equips []*models.Equipment) []*dao.EquipmentBrief {
// 	list := make([]*dao.EquipmentBrief, 0, len(equips))
// 	for _, equip := range equips {
// 		list = append(list, EquipModel2EquipmentBrief(equip))
// 	}
// 	return list
// }

func EquipDao2ModelMap(equipDao []*dao.Equipment, withDetail bool) map[values.EquipId]*models.Equipment {
	mapData := make(map[values.EquipId]*models.Equipment, len(equipDao))
	for _, equipment := range equipDao {
		mapData[equipment.EquipId] = EquipDao2Model(equipment, withDetail)
	}
	return mapData
}

func EquipDao2Models(equipDao []*dao.Equipment, withDetail bool) []*models.Equipment {
	list := make([]*models.Equipment, 0, len(equipDao))
	for _, equipment := range equipDao {
		list = append(list, EquipDao2Model(equipment, withDetail))
	}
	return list
}

func RelicsModel2Dao(relicsModel *models.Relics) *dao.Relics {
	return (*dao.Relics)(relicsModel)
}

func RelicsDao2Model(relicsDao *dao.Relics) *models.Relics {
	return (*models.Relics)(relicsDao)
}

func RelicsModels2Dao(relicsModel []*models.Relics) []*dao.Relics {
	return *(*[]*dao.Relics)(unsafe.Pointer(&relicsModel))
}

func RelicsDao2Models(relicsDao []*dao.Relics) []*models.Relics {
	return *(*[]*models.Relics)(unsafe.Pointer(&relicsDao))
}

func SkillStoneDao2Models(stoneDao []*dao.SkillStone) []*models.SkillStone {
	return *(*[]*models.SkillStone)(unsafe.Pointer(&stoneDao))
}

func TalentRuneDao2Model(runeDao *dao.TalentRune) *models.TalentRune {
	return (*models.TalentRune)(runeDao)
}

func TalentRuneModel2Dao(runeModel *models.TalentRune) *dao.TalentRune {
	return (*dao.TalentRune)(runeModel)
}

func TalentRuneDao2Models(runeDao []*dao.TalentRune) []*models.TalentRune {
	return *(*[]*models.TalentRune)(unsafe.Pointer(&runeDao))
}

func TalentRuneModels2Dao(runeModels []*models.TalentRune) []*dao.TalentRune {
	return *(*[]*dao.TalentRune)(unsafe.Pointer(&runeModels))
}

type EquipQualityNum struct {
	Quality values.Quality
	Num     values.Integer
}

type EquipAffixInfo struct {
	FixedAttr   map[values.AttrId]values.Float
	PercentAttr map[values.AttrId]values.Float
	Buff        []values.HeroBuffId
	Talent      []values.TalentId
}
