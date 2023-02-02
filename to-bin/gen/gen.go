package gen

import (
	"errors"
	"os"
	"strconv"

	"coin-server/common/logger"
	cppbattle "coin-server/common/proto/configgo"
	"coin-server/common/utils"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func LoadRule() {
	r := utils.GetRuleName()
	logger.DefaultLogger.Info("loading latest rule", zap.String("branch", r))
	req := rule.NewRequest()
	if err := req.Get(r, ""); err != nil {
		panic(err)
	}
	v := viper.New()
	version := strconv.Itoa(int(req.Result.Version))
	v.Set(rule_model.VersionKey, version)
	for _, data := range req.Result.Rule {
		v.Set(data.Table, data.Data)
	}
	data := rule_model.NewData(v)
	tables := &cppbattle.Tables{
		KeyValue: ParseKeyValue(data),
		Attr: ParseAttr(data),
		AttrTrans: ParseAttrTrans(data),
		Buff: ParseBuff(data),
		BuffEffect: ParseBuffEffect(data),
		Drop: ParseDrop(data),
		DropLists: ParseDropLists(data),
		Dungeon: ParseDungeon(data),
		InhibitAtk: ParseInhibitAtk(data),
		MapScene: ParseMapScene(data),
		Mechanics: ParseMechanics(data),
		Medicament: ParseMedicament(data),
		Monster: ParseMonster(data),
		MonsterGroup: ParseMonsterGroup(data),
		Robot: ParseRobot(data),
		RoguelikeArtifact: ParseRoguelikeArtifact(data),
		RoguelikeDungeon: ParseRoguelikeDungeon(data),
		RoleLv: ParseRoleLv(data),
		RoleReachDungeon: ParseRoleReachDungeon(data),
		RowHero: ParseRowHero(data),
		Skill: ParseSkill(data),
		Summoned: ParseSummoned(data),
		TempBag: ParseTempBag(data),
	}
	outFilePath := "./table.bin.txt"
	_ = os.Remove(outFilePath)
	outF, err := os.Create(outFilePath)
	defer outF.Close()
	if err != nil {
		panic(err)
	}
	tableBytes, err := tables.Marshal()
	if err != nil {
		panic(err)
	}
	outF.Write(tableBytes)
	outF.Sync()
}
func ParseKeyValue(data *rule_model.Data) map[string]*cppbattle.KeyValue {
	list := make([]*cppbattle.KeyValue, 0)
	if err := data.UnmarshalKey("key_value", &list); err != nil {
		panic(errors.New("parse table key_value err:\n" + err.Error()))
	}
	m := make(map[string]*cppbattle.KeyValue)
	for idx, l := range list {
		m[l.Key] = list[idx]
	}
	return m
}
func ParseAttr(data *rule_model.Data) map[int64]*cppbattle.Attr {
	list := make([]*cppbattle.Attr, 0)
	if err := data.UnmarshalKey("attr", &list); err != nil {
		panic(errors.New("parse table attr err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Attr)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseAttrTrans(data *rule_model.Data) map[int64]*cppbattle.AttrTrans {
	list := make([]*cppbattle.AttrTrans, 0)
	if err := data.UnmarshalKey("attr_trans", &list); err != nil {
		panic(errors.New("parse table attr_trans err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.AttrTrans)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseBuff(data *rule_model.Data) map[int64]*cppbattle.Buff {
	list := make([]*cppbattle.Buff, 0)
	if err := data.UnmarshalKey("buff", &list); err != nil {
		panic(errors.New("parse table buff err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Buff)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseBuffEffect(data *rule_model.Data) map[int64]*cppbattle.BuffEffect {
	list := make([]*cppbattle.BuffEffect, 0)
	if err := data.UnmarshalKey("buff_effect", &list); err != nil {
		panic(errors.New("parse table buff_effect err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.BuffEffect)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseDrop(data *rule_model.Data) map[int64]*cppbattle.Drop {
	list := make([]*cppbattle.Drop, 0)
	if err := data.UnmarshalKey("drop", &list); err != nil {
		panic(errors.New("parse table drop err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Drop)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	mChild := ParseDropMini(data)
	for _, child := range mChild {
		if _, exist := m[child.DropId]; exist {
			if m[child.DropId].DropMini == nil {
				m[child.DropId].DropMini = map[int64]*cppbattle.DropMini{}
			}
			m[child.DropId].DropMini[child.Id] = child
		}
	}
	return m
}


	func ParseDropMini(data *rule_model.Data) []*cppbattle.DropMini {
	list := make([]*cppbattle.DropMini, 0)
	if err := data.UnmarshalKey("drop_mini", &list); err != nil {
		panic(errors.New("parse table drop_mini err:\n" + err.Error()))
	}
	return list
}
func ParseDropLists(data *rule_model.Data) map[int64]*cppbattle.DropLists {
	list := make([]*cppbattle.DropLists, 0)
	if err := data.UnmarshalKey("drop_lists", &list); err != nil {
		panic(errors.New("parse table drop_lists err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.DropLists)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	mChild := ParseDropListsMini(data)
	for _, child := range mChild {
		if _, exist := m[child.DropListsId]; exist {
			if m[child.DropListsId].DropListsMini == nil {
				m[child.DropListsId].DropListsMini = map[int64]*cppbattle.DropListsMini{}
			}
			m[child.DropListsId].DropListsMini[child.Id] = child
		}
	}
	return m
}


	func ParseDropListsMini(data *rule_model.Data) []*cppbattle.DropListsMini {
	list := make([]*cppbattle.DropListsMini, 0)
	if err := data.UnmarshalKey("drop_lists_mini", &list); err != nil {
		panic(errors.New("parse table drop_lists_mini err:\n" + err.Error()))
	}
	return list
}
func ParseDungeon(data *rule_model.Data) map[int64]*cppbattle.Dungeon {
	list := make([]*cppbattle.Dungeon, 0)
	if err := data.UnmarshalKey("dungeon", &list); err != nil {
		panic(errors.New("parse table dungeon err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Dungeon)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseInhibitAtk(data *rule_model.Data) map[int64]*cppbattle.InhibitAtk {
	list := make([]*cppbattle.InhibitAtk, 0)
	if err := data.UnmarshalKey("inhibit_atk", &list); err != nil {
		panic(errors.New("parse table inhibit_atk err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.InhibitAtk)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseMapScene(data *rule_model.Data) map[int64]*cppbattle.MapScene {
	list := make([]*cppbattle.MapScene, 0)
	if err := data.UnmarshalKey("map_scene", &list); err != nil {
		panic(errors.New("parse table map_scene err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.MapScene)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseMechanics(data *rule_model.Data) map[int64]*cppbattle.Mechanics {
	list := make([]*cppbattle.Mechanics, 0)
	if err := data.UnmarshalKey("mechanics", &list); err != nil {
		panic(errors.New("parse table mechanics err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Mechanics)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseMedicament(data *rule_model.Data) map[int64]*cppbattle.Medicament {
	list := make([]*cppbattle.Medicament, 0)
	if err := data.UnmarshalKey("medicament", &list); err != nil {
		panic(errors.New("parse table medicament err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Medicament)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseMonster(data *rule_model.Data) map[int64]*cppbattle.Monster {
	list := make([]*cppbattle.Monster, 0)
	if err := data.UnmarshalKey("monster", &list); err != nil {
		panic(errors.New("parse table monster err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Monster)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseMonsterGroup(data *rule_model.Data) map[int64]*cppbattle.MonsterGroup {
	list := make([]*cppbattle.MonsterGroup, 0)
	if err := data.UnmarshalKey("monster_group", &list); err != nil {
		panic(errors.New("parse table monster_group err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.MonsterGroup)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseRobot(data *rule_model.Data) map[int64]*cppbattle.Robot {
	list := make([]*cppbattle.Robot, 0)
	if err := data.UnmarshalKey("robot", &list); err != nil {
		panic(errors.New("parse table robot err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Robot)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseRoguelikeArtifact(data *rule_model.Data) map[int64]*cppbattle.RoguelikeArtifact {
	list := make([]*cppbattle.RoguelikeArtifact, 0)
	if err := data.UnmarshalKey("roguelike_artifact", &list); err != nil {
		panic(errors.New("parse table roguelike_artifact err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.RoguelikeArtifact)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseRoguelikeDungeon(data *rule_model.Data) map[int64]*cppbattle.RoguelikeDungeon {
	list := make([]*cppbattle.RoguelikeDungeon, 0)
	if err := data.UnmarshalKey("roguelike_dungeon", &list); err != nil {
		panic(errors.New("parse table roguelike_dungeon err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.RoguelikeDungeon)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	mChild := ParseRoguelikeDungeonRoom(data)
	for _, child := range mChild {
		if _, exist := m[child.RoguelikeDungeonId]; exist {
			if m[child.RoguelikeDungeonId].RoguelikeDungeonRoom == nil {
				m[child.RoguelikeDungeonId].RoguelikeDungeonRoom = map[int64]*cppbattle.RoguelikeDungeonRoom{}
			}
			m[child.RoguelikeDungeonId].RoguelikeDungeonRoom[child.Id] = child
		}
	}
	return m
}


	func ParseRoguelikeDungeonRoom(data *rule_model.Data) []*cppbattle.RoguelikeDungeonRoom {
	list := make([]*cppbattle.RoguelikeDungeonRoom, 0)
	if err := data.UnmarshalKey("roguelike_dungeon_room", &list); err != nil {
		panic(errors.New("parse table roguelike_dungeon_room err:\n" + err.Error()))
	}
	return list
}
func ParseRoleLv(data *rule_model.Data) map[int64]*cppbattle.RoleLv {
	list := make([]*cppbattle.RoleLv, 0)
	if err := data.UnmarshalKey("role_lv", &list); err != nil {
		panic(errors.New("parse table role_lv err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.RoleLv)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseRoleReachDungeon(data *rule_model.Data) map[int64]*cppbattle.RoleReachDungeon {
	list := make([]*cppbattle.RoleReachDungeon, 0)
	if err := data.UnmarshalKey("role_reach_dungeon", &list); err != nil {
		panic(errors.New("parse table role_reach_dungeon err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.RoleReachDungeon)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseRowHero(data *rule_model.Data) map[int64]*cppbattle.RowHero {
	list := make([]*cppbattle.RowHero, 0)
	if err := data.UnmarshalKey("row_hero", &list); err != nil {
		panic(errors.New("parse table row_hero err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.RowHero)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseSkill(data *rule_model.Data) map[int64]*cppbattle.Skill {
	list := make([]*cppbattle.Skill, 0)
	if err := data.UnmarshalKey("skill", &list); err != nil {
		panic(errors.New("parse table skill err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Skill)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseSummoned(data *rule_model.Data) map[int64]*cppbattle.Summoned {
	list := make([]*cppbattle.Summoned, 0)
	if err := data.UnmarshalKey("summoned", &list); err != nil {
		panic(errors.New("parse table summoned err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.Summoned)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

func ParseTempBag(data *rule_model.Data) map[int64]*cppbattle.TempBag {
	list := make([]*cppbattle.TempBag, 0)
	if err := data.UnmarshalKey("temp_bag", &list); err != nil {
		panic(errors.New("parse table temp_bag err:\n" + err.Error()))
	}
	m := make(map[int64]*cppbattle.TempBag)
	for idx, l := range list {
		m[l.Id] = list[idx]
	}
	return m
}

