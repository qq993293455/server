package service

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/gvgguild"
	"coin-server/common/proto/models"
	"coin-server/common/skiplist"
	"coin-server/common/utils"
)

const (
	GuildFunctionGVG = 5
)

func GetChatRoomID(groupId int64) string {
	return "gvgguild_" + strconv.Itoa(int(groupId))
}

type GuildAvgPowerInfo struct {
	AvgPower int64
	GuildId  string
	Roles    []*dao.Role
	Level    int64
}

type FightInfo struct {
	Id               int64     `json:"id"`
	GroupId          int64     `json:"group_id"`
	AttackGuildId    string    `json:"attack_guild_id"`
	Attacker         string    `json:"attacker"`
	DefendGuildId    string    `json:"defend_guild_id"`
	Defender         string    `json:"defender"`
	IsBuilder        int64     `json:"is_builder"`
	Blood            int64     `json:"blood"`
	IsWin            int64     `json:"is_win"`
	CreateTime       time.Time `json:"create_time"`
	PersonalScoreAdd int64     `json:"personal_score_add"`
	GuildScoreAdd    int64     `json:"guild_score_add"`
	BuildId          int64     `json:"build_id"` //如果攻击建筑。
}

type ChatFightInfo struct {
	AttackGuildId string `json:"-"`
	Attacker      string `json:"-"`
	DefendGuildId string `json:"-"`
	Defender      string `json:"-"`
	GroupId       int64  `json:"-"`

	AttackGuildName  string `json:"attack_guild_name"`
	AttackerName     string `json:"attacker_name"`
	DefendGuildName  string `json:"defend_guild_name"`
	DefenderName     string `json:"defender_name"`
	IsBuilder        int64  `json:"is_builder"`
	Blood            int64  `json:"blood"`
	IsWin            int64  `json:"is_win"`
	PersonalScoreAdd int64  `json:"personal_score_add"`
	GuildScoreAdd    int64  `json:"guild_score_add"`
	BuildId          int64  `json:"build_id"` //如果攻击建筑。
}

type GuildInfo struct {
	Id              string      `json:"id"`
	GroupId         int64       `json:"group_id"`
	Data            []BuildInfo `json:"-"`
	FlagGuildId     string      `json:"flag_guild_id"` //标记
	Score           int64       `json:"score"`
	LastScoreChange int64       `json:"last_score_change"`
	AttackPer       int64       `json:"attack_per"`      //攻击加成万分比
	DefendPer       int64       `json:"defend_per"`      //防御加成万分比
	LastFightInfo   int64       `json:"last_fight_info"` //最后一次被攻击的信息
	IsSettle        int64       `json:"-"`
	AddBuffIds      []int64     `json:"add_buff_ids"`  // 获取的buff
	AddExtScore     int64       `json:"add_ext_score"` // 额外的得分加成
	ReliveTime      int64       `json:"reliveTime"`    // 英灵殿被摧毁后的复活时间
}

type BuildInfo struct {
	Id           int64       `json:"id"`
	Blood        int64       `json:"blood"`    //建筑血量
	Priority     int64       `json:"priority"` //建筑属性，需要先挑战小的
	MaxRoleCount int64       `json:"max_role_count"`
	Roles        []BuildRole `json:"roles"`
	MaxBlood     int64       `json:"max_blood"` // 最大血量
}

type BuildRole struct {
	RoleId          string  `json:"role_id"`
	IsHead          bool    `json:"is_head"` // 带头守将
	KillCount       int64   `json:"kill_count"`
	Score           int64   `json:"score"`
	BuildHurt       int64   `json:"build_hurt"`
	NextAddTimes    int64   `json:"last_attack_time"` // 上次增加攻击次数的时间
	CanAttackCount  int64   `json:"can_attack_count"` // 剩余可攻击次数
	IsDeath         bool    `json:"is_death"`         // 是否已战败
	GuildId         string  `json:"guild_id"`         // 匹配时工会ID
	IsFighting      bool    `json:"is_fighting"`      //是否在战斗中
	StartFightTime  int64   `json:"start_fight_time"` // 如果在战斗中。战斗的开始时间
	Attacker        string  `json:"attacker"`         // 正在攻击自己的玩家
	FightLog        []int64 `json:"fight_log"`        // 战斗记录
	BuildId         int64   `json:"build_id"`
	ScoreChangeTime int64   `json:"sct"` //分数最后变化的时候
}

type GroupInfo struct {
	Id         int64              `json:"id"`
	Infos      []GuildInfo        `json:"infos"`
	CreateTime time.Time          `json:"create_time"`
	sl         *skiplist.SkipList // 最大可能才1000个元素。
	fiMap      map[int64]*FightInfo
}

type BuildRoleComparator struct {
}

func (BuildRoleComparator) Equal(a, b interface{}) bool {
	v1, v2 := a.(*models.RankValue), b.(*models.RankValue)
	if len(v1.Extra) != len(v2.Extra) {
		return false
	}
	if v1.Extra != nil && v2.Extra != nil {
		for k, v := range v1.Extra {
			if v2.Extra[k] != v {
				return false
			}
		}
	}
	if v1.OwnerId == v2.OwnerId && v1.Value1 == v2.Value1 && v1.Value2 == v2.Value2 &&
		v1.Value3 == v2.Value3 && v1.Value4 == v2.Value4 {
		return true
	}
	return false
}

func (BuildRoleComparator) Less(a1, b1 interface{}) bool {
	a, b := a1.(*models.RankValue), b1.(*models.RankValue)
	if a.Value1 != b.Value1 {
		return a.Value1 > b.Value1
	}

	// 积分相同比较时间，时间小的在前面
	if a.CreatedAt != b.CreatedAt {
		return a.CreatedAt < b.CreatedAt
	}
	return a.OwnerId < b.OwnerId
}

func newSkipList() *skiplist.SkipList {
	sl := skiplist.NewSkipList(BuildRoleComparator{})
	return sl
}

func SortGuildInfos(data []GuildInfo) {
	sort.Slice(data, func(i, j int) bool {
		gi := &data[i]
		gj := &data[j]
		if gi.Score > gj.Score {
			return true
		} else if gi.Score < gj.Score {
			return false
		}
		if gi.LastScoreChange < gj.LastScoreChange {
			return true
		} else if gi.LastScoreChange > gj.LastScoreChange {
			return false
		}
		return gi.Id < gj.Id
	})
}

type GuildInfoTempInfo struct {
	*gvgguild.GuildGVG_GuildInfo
	Score           int64
	LastScoreChange int64
}

const (
	maxGuildNum      = 10000000
	onceFightSeconds = 180 + 60
)

//func CurrActiveID() int64 {
//	return utils.GetCurrWeek() * maxGuildNum
//}

func NextActiveID(t time.Time) int64 {
	now := t.Add(time.Hour * 24 * 7).UTC()
	return ActiveIDWithTime(now)
}

func ActiveIDWithTime(t time.Time) int64 {
	return utils.GetWeekWithTime(t) * maxGuildNum
}

type CheckSignInfo struct {
	Id       string
	IsSignup bool
	Nickname string
}

type CheckAndSetSignInfo struct {
	Id            string
	NickName      string
	SignupSuccess bool
}

type DeleteSignupInfo struct {
	Id string
}

type BuildInfoSave struct {
	GroupId int64
	GuildId string
	BuildId int64
	Data    json.RawMessage
}

type GuildInfoSave struct {
	GroupId int64
	GuildId string
	Data    json.RawMessage
}

func CreateGuildInfoSave(gi *GuildInfo) *GuildInfoSave {
	gis := &GuildInfoSave{
		GroupId: gi.GroupId,
		GuildId: gi.Id,
		Data:    nil,
	}
	var e error
	gis.Data, e = json.Marshal(gi)
	utils.Must(e)
	return gis
}

func CreateBuildInfoSave(groupId int64, guildId string, bi *BuildInfo) *BuildInfoSave {
	if bi == nil {
		return nil
	}
	bis := &BuildInfoSave{
		GroupId: groupId,
		GuildId: guildId,
		BuildId: bi.Id,
	}
	var e error
	bis.Data, e = json.Marshal(bi)
	utils.Must(e)
	return bis
}

type FightingSave FightInfo
