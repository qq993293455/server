package service

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/gopool"
	"coin-server/common/im"
	"coin-server/common/proto/dao"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/safego"
	"coin-server/common/timer"
	"coin-server/common/utils"
	guildDao "coin-server/game-server/service/guild/dao"
	userDao "coin-server/game-server/service/user/db"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var databaseCreate = "CREATE DATABASE IF NOT EXISTS gvg DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_general_ci;"

var createTableMap = map[string]string{
	"signup": `CREATE TABLE gvg.signup  (
 id int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
 act_id bigint NOT NULL COMMENT '本次活动ID',
 guild_id varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '工会id',
 nick_name varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '报名人的昵称',
 is_match bigint NOT NULL DEFAULT 0 COMMENT '是否已经匹配过了',
 create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
 PRIMARY KEY (id) USING BTREE,
 UNIQUE INDEX guild_id(guild_id) USING BTREE COMMENT '唯一键',
 INDEX act_id(act_id DESC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;`,
	"group": `CREATE TABLE gvg.group  (
  id bigint UNSIGNED NOT NULL COMMENT '主键',
  guilds json NOT NULL COMMENT '当前组的工会列表',
  is_settle bigint NOT NULL DEFAULT 0 COMMENT '是否完成结算',
  create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;`,
	"reward": `CREATE TABLE gvg.reward  (
  id bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  activity_id bigint NOT NULL COMMENT '本次活动处于什么状态,0 报名中,1 匹配阶段，2 战斗阶段，3 结算阶段',
  role_id varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户ID',
  send_guild_reward bigint NOT NULL DEFAULT 0 COMMENT '是否发放工会奖励',
  send_personal_reward bigint NOT NULL DEFAULT 0 COMMENT '是否发送个人奖励',
  create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  PRIMARY KEY (id) USING BTREE,
  UNIQUE INDEX activity_id(activity_id ASC, role_id ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '是否发放奖品' ROW_FORMAT = DYNAMIC;`,
	"guild": `CREATE TABLE gvg.guild  (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  group_id bigint NOT NULL COMMENT '分组ID',
  guild_id varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '工会id',
  guild_data json NOT NULL COMMENT '工会数据',
  create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_settle bigint NOT NULL DEFAULT 0 COMMENT '是否已经结算',
  updated_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  PRIMARY KEY (id) USING BTREE,
  UNIQUE INDEX group_guild_index(group_id DESC, guild_id DESC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '工会信息' ROW_FORMAT = Dynamic;`,
	"build": `CREATE TABLE gvg.build  (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  group_id bigint NOT NULL COMMENT '组id',
  guild_id varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '工会id',
  build_id bigint NOT NULL COMMENT '建筑ID',
  build_info json NOT NULL COMMENT '建筑信息',
  PRIMARY KEY (id) USING BTREE,
  UNIQUE INDEX build_group_id_index(group_id DESC, guild_id DESC, build_id DESC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;`,
	"fighting": `CREATE TABLE gvg.fighting  (
  id bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  group_id bigint NOT NULL COMMENT '哪个分组',
  attack_guild_id varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '攻击工会',
  attacker varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '攻击者',
  defend_guild_id varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '防御工会，可能是建筑',
  defender varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '防御者',
  is_builder bigint NOT NULL DEFAULT 0 COMMENT '是否是攻击的建筑',
  blood bigint NOT NULL COMMENT '造成血量',
  is_win bigint NOT NULL COMMENT '是否攻击胜利',
  create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  personal_score_add bigint NOT NULL DEFAULT 0 COMMENT '个人积分增加',
  guild_score_add bigint NOT NULL COMMENT '工会积分增加',
  build_id bigint NOT NULL COMMENT '建筑id',
  PRIMARY KEY (id) USING BTREE,
  INDEX attacker(attack_guild_id ASC) USING BTREE,
  INDEX defender(defend_guild_id ASC) USING BTREE,
  INDEX group_index(group_id ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;`,
}

func (this_ *Service) delOldTable() {
	this_.log.Info("start drop old table")
	startTime := time.Now().UTC()
	defer func() { this_.log.Info("end drop old table", zap.Duration("cost", time.Now().UTC().Sub(startTime))) }()

	var tableNames []string
	err := this_.mysql.Query(func(rows *sqlx.Rows) error {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return err
		}
		tableNames = append(tableNames, name)
		return nil
	}, "SELECT table_name FROM information_schema.TABLES WHERE table_schema='gvg'")
	if err != nil {
		this_.log.Warn("delete gvg table failed", zap.Error(err))
	}
	now := time.Now().UTC()
	for _, v := range tableNames {
		strS := strings.Split(v, "_")
		if len(strS) == 2 {
			t, err := time.ParseInLocation("20060102150405", strS[1], time.UTC)
			if err != nil {
				this_.log.Warn("time.ParseInLocation error", zap.Error(err), zap.String("value", v))
				continue
			}
			if now.Sub(t).Hours() >= 30*24 { // 超过1个月，删
				_, _, err := this_.mysql.Exec("DROP TABLE gvg." + v)
				if err != nil {
					this_.log.Warn("mysql drop table error", zap.Error(err), zap.String("table", v))
				}
			}
		}
	}
}

func (this_ *Service) StartDropOldTable() {
	this_.delOldTable()
	timer.AfterFunc(time.Hour*6, func() {
		safego.RecoverWithLogger(this_.log)
		this_.delOldTable()
	})
}

func (this_ *Service) checkTable() {
	_, _, err := this_.mysql.Exec(databaseCreate)
	if err != nil {
		panic(err)
	}
	existsSignup := 0
	err, _ = this_.mysql.QueryRow(func(row *sqlx.Row) error {
		return row.Scan(&existsSignup)
	}, "SELECT COUNT(*) FROM information_schema.TABLES WHERE table_name='signup' AND table_schema='gvg'")
	if err != nil {
		panic(err)
	}
	if existsSignup != 1 { //创建报名表
		_, _, err = this_.mysql.Exec(createTableMap["signup"])
		if err != nil {
			panic(err)
		}
	}
	existsSignup = 0
	err, _ = this_.mysql.QueryRow(func(row *sqlx.Row) error {
		return row.Scan(&existsSignup)
	}, "SELECT COUNT(*) FROM information_schema.TABLES WHERE table_name='group' AND table_schema='gvg'")
	if err != nil {
		panic(err)
	}
	if existsSignup != 1 {
		_, _, err = this_.mysql.Exec(createTableMap["group"])
		if err != nil {
			panic(err)
		}
	}

	existsSignup = 0
	err, _ = this_.mysql.QueryRow(func(row *sqlx.Row) error {
		return row.Scan(&existsSignup)
	}, "SELECT COUNT(*) FROM information_schema.TABLES WHERE table_name='reward' AND table_schema='gvg'")
	if err != nil {
		panic(err)
	}
	if existsSignup != 1 {
		_, _, err = this_.mysql.Exec(createTableMap["reward"])
		if err != nil {
			panic(err)
		}
	}

	existsSignup = 0
	err, _ = this_.mysql.QueryRow(func(row *sqlx.Row) error {
		return row.Scan(&existsSignup)
	}, "SELECT COUNT(*) FROM information_schema.TABLES WHERE table_name='guild' AND table_schema='gvg'")
	if err != nil {
		panic(err)
	}
	if existsSignup != 1 {
		_, _, err = this_.mysql.Exec(createTableMap["guild"])
		if err != nil {
			panic(err)
		}
	}

	existsSignup = 0
	err, _ = this_.mysql.QueryRow(func(row *sqlx.Row) error {
		return row.Scan(&existsSignup)
	}, "SELECT COUNT(*) FROM information_schema.TABLES WHERE table_name='build' AND table_schema='gvg'")
	if err != nil {
		panic(err)
	}
	if existsSignup != 1 {
		_, _, err = this_.mysql.Exec(createTableMap["build"])
		if err != nil {
			panic(err)
		}
	}

	existsSignup = 0
	err, _ = this_.mysql.QueryRow(func(row *sqlx.Row) error {
		return row.Scan(&existsSignup)
	}, "SELECT COUNT(*) FROM information_schema.TABLES WHERE table_name='fighting' AND table_schema='gvg'")
	if err != nil {
		panic(err)
	}
	if existsSignup != 1 {
		_, _, err = this_.mysql.Exec(createTableMap["fighting"])
		if err != nil {
			panic(err)
		}
	}
}

func (this_ *Service) init() {
	this_.checkTable()
	this_.loadData()
}

func (this_ *Service) loadData() {
	now := time.Now().UTC()
	this_.activeId = ActiveIDWithTime(now)
	ac := this_.GetConfig()
	thisWeekBeginTime := timer.BeginOfWeek(now)
	this_.activeEndTime = thisWeekBeginTime.Add(time.Duration((ac.endDay-1)*86400+ac.closeDayTime*3600) * time.Second).Unix() // 这个星期的活动开始时间
	this_.activeStartTime = thisWeekBeginTime.Add(time.Duration((ac.startDay-1)*86400+ac.startDayTime*3600) * time.Second).Unix()
	if now.Unix() > this_.activeEndTime { // 进入了下一个阶段
		this_.activeId = NextActiveID(now)
		thisWeekBeginTime := timer.BeginOfWeek(now.Add(time.Hour * 24 * 7))
		this_.activeEndTime = thisWeekBeginTime.Add(time.Duration((ac.endDay-1)*86400+ac.closeDayTime*3600) * time.Second).Unix() // 这个星期的活动开始时间
		this_.activeStartTime = thisWeekBeginTime.Add(time.Duration((ac.startDay-1)*86400+ac.startDayTime*3600) * time.Second).Unix()
	}
	// 获取最近一次的活动ID
	var oldId int64
	err := this_.mysql.Query(func(rows *sqlx.Rows) error {
		err := rows.Scan(&oldId)
		return err
	}, "SELECT id FROM gvg.group ORDER BY id asc limit 1")
	if err != nil {
		panic(err)
	}
	oldId = (oldId / 10000000) * 10000000
	// 无脑处理。直接先把所有数据加载到内存中
	err = this_.mysql.Query(func(rows *sqlx.Rows) error {
		var si signupInfo
		var guildId string
		var isMatch int64
		e := rows.Scan(&guildId, &si.NickName, &isMatch, &si.SignTime)
		if e != nil {
			return e
		}
		si.IsMatch = isMatch == 1
		this_.signupMap[guildId] = si
		return nil
	}, "SELECT guild_id,nick_name,is_match,create_time FROM gvg.signup")
	if err != nil {
		panic(err)
	}
	err = this_.mysql.Query(func(rows *sqlx.Rows) error {
		var id int64
		e := rows.Scan(&id)
		if e != nil {
			return e
		}
		this_.groupMap[id] = &GroupInfo{Id: id, sl: newSkipList(), fiMap: map[int64]*FightInfo{}}
		return nil
	}, "SELECT id FROM gvg.group")
	if err != nil {
		panic(err)
	}
	err = this_.mysql.Query(func(rows *sqlx.Rows) error {
		var groupId int64
		var guildId string
		var jsonData json.RawMessage
		e := rows.Scan(&groupId, &guildId, &jsonData)
		if e != nil {
			return e
		}
		var guildInfo GuildInfo
		e = json.Unmarshal(jsonData, &guildInfo)
		if e != nil {
			return e
		}
		gi := this_.groupMap[groupId]
		gi.Infos = append(gi.Infos, guildInfo)
		this_.guildGroupInfo[guildId] = groupId
		return nil
	}, "SELECT group_id,guild_id,guild_data FROM gvg.guild")
	if err != nil {
		panic(err)
	}
	err = this_.mysql.Query(func(rows *sqlx.Rows) error {
		var groupId int64
		var guildId string
		var buildId int64
		var jsonData json.RawMessage
		e := rows.Scan(&groupId, &guildId, &buildId, &jsonData)
		if e != nil {
			return e
		}
		var buildInfo BuildInfo
		e = json.Unmarshal(jsonData, &buildInfo)
		if e != nil {
			return e
		}
		gi := this_.groupMap[groupId]
		for i := range gi.Infos {
			if gi.Infos[i].Id == guildId {
				gi.Infos[i].Data = append(gi.Infos[i].Data, buildInfo)
				break
			}
		}
		for i := range buildInfo.Roles {
			role := &buildInfo.Roles[i]
			if role.Score > 0 {
				gi.sl.Insert(role.RoleId, &models.RankValue{
					OwnerId:   role.RoleId,
					Value1:    role.Score,
					CreatedAt: role.ScoreChangeTime,
				})
			}
			this_.userGroupInfo[role.RoleId] = role
		}
		return nil
	}, "SELECT group_id,guild_id,build_id,build_info FROM gvg.build")
	if err != nil {
		panic(err)
	}

	err = this_.mysql.Query(func(rows *sqlx.Rows) error {
		f := &FightInfo{}
		err := rows.Scan(&f.Id, &f.GroupId, &f.AttackGuildId, &f.Attacker, &f.DefendGuildId, &f.Defender,
			&f.IsBuilder, &f.Blood, &f.IsWin, &f.CreateTime, &f.PersonalScoreAdd, &f.GuildScoreAdd, &f.BuildId)
		if err != nil {
			return err
		}
		group, ok := this_.groupMap[f.GroupId]
		if ok {
			group.fiMap[f.Id] = f
		}
		return nil
	}, "SELECT id,group_id,attack_guild_id,attacker,defend_guild_id,defender,is_builder,blood,is_win,create_time,personal_score_add,guild_score_add,build_id FROM gvg.fighting")

	if oldId != 0 && oldId != this_.activeId { //如果不相等。需要做结算
		// 找出没有战斗结果的分组删除
		delKs := make([]int64, 0, len(this_.groupMap))
		for k, v := range this_.groupMap {
			if len(v.fiMap) == 0 {
				delKs = append(delKs, k)
			}
		}
		for _, v := range delKs {
			delete(this_.groupMap, v)
		}
		// 剩下的是有战斗的分组,开始结算
		this_.activeId = oldId
		this_.startSettle()
	}
	this_.monitoringStatus()
}

func (this_ *Service) match() {
	this_.log.Info("gvg_status: 开始匹配")
	// 先获取战斗力
	l := len(this_.signupMap)
	guildIds := make([]string, 0, l)
	for k := range this_.signupMap {
		if !this_.signupMap[k].IsMatch {
			guildIds = append(guildIds, k)
		}
	}
	// 获取成员最大战力的平均值
	avgPowerSlice := make([]GuildAvgPowerInfo, 0, len(guildIds))
	for _, guildId := range guildIds {
		c := ctx.GetContext()
		members, err := guildDao.NewGuildMember(guildId).Get(c)
		if err != nil {
			this_.log.Warn("读取工会成员数据出错,丢弃此工会", zap.String("guildId", guildId), zap.Error(err))
			continue
		}
		roles := make([]*dao.Role, 0, len(members))
		for _, v := range members {
			roles = append(roles, &dao.Role{RoleId: v.RoleId})
		}
		roles, err = userDao.GetMultiRole(c, roles)
		if err != nil {
			this_.log.Warn("读取工会成员Role数据出错,丢弃此工会", zap.String("guildId", guildId), zap.Error(err))
			continue
		}
		if len(roles) > 0 {
			g, err := guildDao.NewGuild(guildId).Get(c)
			if err != nil {
				this_.log.Warn("读取工会数据出错,丢弃此工会", zap.String("guildId", guildId), zap.Error(err))
				continue
			}
			totalPower := int64(0)
			for _, v := range roles {
				totalPower += v.HighestPower
			}
			avgPower := totalPower / int64(len(roles))
			avgPowerSlice = append(avgPowerSlice, GuildAvgPowerInfo{
				AvgPower: avgPower,
				GuildId:  guildId,
				Roles:    roles,
				Level:    g.Level,
			})
		}
	}
	// 开始分组
	sort.Slice(avgPowerSlice, func(i, j int) bool {
		return avgPowerSlice[i].AvgPower < avgPowerSlice[j].AvgPower
	})
	r := rule.MustGetReader(nil)
	buildCnf := r.GuildContendbuild.List()
	guildContendChallengeNum, ok := r.KeyValue.GetInt64("GuildContendChallengeNum") // 公会GVG：活动开启默认挑战次数
	utils.MustTrue(ok)
	guildContendEveryIntegral, ok := r.KeyValue.GetInt64("GuildContendEveryIntegral") // 公会GVG：公会最初拥有的积分值
	utils.MustTrue(ok)
	// guildContendChallengeNumMax, ok := r.KeyValue.GetInt64("GuildContendChallengeNumMax") // 公会GVG：最大可积累挑战次数
	groupId := this_.activeId
	for len(avgPowerSlice) > 0 {
		thisGroupGuildNum := 10
		if len(avgPowerSlice) <= 15 && len(avgPowerSlice) > 10 {
			thisGroupGuildNum = 7
		}
		if len(avgPowerSlice) <= 10 {
			thisGroupGuildNum = len(avgPowerSlice)
		}
		gi := &GroupInfo{
			Id:         groupId,
			Infos:      make([]GuildInfo, thisGroupGuildNum),
			CreateTime: timer.Now(),
			sl:         newSkipList(),
			fiMap:      map[int64]*FightInfo{},
		}
		roomId := GetChatRoomID(groupId)
		start := len(avgPowerSlice) - 1
		chatRoles := make([]string, 0, 512)
		for i := 0; i < thisGroupGuildNum; i++ {
			gap := &avgPowerSlice[start]
			gii := &gi.Infos[i]
			gii.Id = gap.GuildId
			gii.GroupId = groupId
			gii.Score = guildContendEveryIntegral
			gii.LastScoreChange = time.Now().UTC().Unix()
			for index := range buildCnf {
				v := &buildCnf[index]
				maxBlood := v.BuildHp + gap.Level*v.BuildHpUp
				gii.Data = append(gii.Data, BuildInfo{
					Id:           v.Id,
					Blood:        maxBlood,
					Priority:     v.BuildPriority,
					MaxRoleCount: (gap.Level-1)*v.NumAdd + v.DefenseNum,
					Roles:        nil,
					MaxBlood:     maxBlood,
				})
			}
			sort.Slice(gii.Data, func(i, j int) bool {
				return gii.Data[i].Id < gii.Data[j].Id
			})
			roles := gap.Roles

			// 先把每个建筑依次丢入战力最高的人
			for i := range gii.Data {
				if len(roles) > 0 {
					role := roles[len(roles)-1]
					gii.Data[i].Roles = append(gii.Data[i].Roles, BuildRole{
						RoleId:         role.RoleId,
						IsHead:         true,
						KillCount:      0,
						Score:          0,
						BuildHurt:      0,
						NextAddTimes:   0,
						CanAttackCount: guildContendChallengeNum,
						IsDeath:        false,
						GuildId:        gii.Id,
						BuildId:        gii.Data[i].Id,
					})
					roles = roles[:len(roles)-1]
					chatRoles = append(chatRoles, role.RoleId)
				} else {
					break
				}
			}
			// 再依次把每个建筑填满
			if len(roles) > 0 {
				for i := range gii.Data {
					for gii.Data[i].MaxRoleCount < int64(len(gii.Data[i].Roles)) {
						if len(roles) > 0 {
							role := roles[len(roles)-1]
							gii.Data[i].Roles = append(gii.Data[i].Roles, BuildRole{
								RoleId:         role.RoleId,
								IsHead:         true,
								KillCount:      0,
								Score:          0,
								BuildHurt:      0,
								NextAddTimes:   0,
								CanAttackCount: guildContendChallengeNum,
								IsDeath:        false,
								GuildId:        gii.Id,
								BuildId:        gii.Data[i].Id,
							})
							roles = roles[:len(roles)-1]
							chatRoles = append(chatRoles, role.RoleId)
						} else {
							break
						}
					}
				}
			}

			for ii := range gii.Data {
				for rix := range gii.Data[ii].Roles {
					ri := &gii.Data[ii].Roles[rix]
					this_.userGroupInfo[ri.RoleId] = ri
				}
			}
			this_.guildGroupInfo[gii.Id] = groupId
			start--
		}
		// 入库
		this_.groupInfoSaveToDB(gi)
		this_.groupMap[groupId] = gi
		gopool.Submit(func() {
			chatErr := im.DefaultClient.JoinRoom(context.Background(), &im.RoomRole{
				RoomID:  roomId,
				RoleIDs: chatRoles,
			})
			if chatErr != nil {
				this_.log.Error("创建聊天室失败", zap.Error(chatErr), zap.String("chatId", roomId))
			}
		})

		groupId++
		avgPowerSlice = avgPowerSlice[:len(avgPowerSlice)-thisGroupGuildNum]
	}
	// 分组结束

	this_.SetGVGStatus(StatusFighting) // 匹配完了进入战斗状态
	this_.log.Info("匹配结束。进入战斗状态", zap.Int64("activityId", this_.activeId))
}

func (this_ *Service) groupInfoSaveToDB(gi *GroupInfo) {
	guildIds := make([]string, 0, len(gi.Infos))
	for i := range gi.Infos {
		gii := &gi.Infos[i]
		guildIds = append(guildIds, gii.Id)
	}
	jsonData, e := json.Marshal(guildIds)
	utils.Must(e)
	err := this_.mysql.WithTxx(func(tx *sqlx.Tx) error {
		_, err := tx.Exec("INSERT INTO gvg.group(id,guilds) VALUES (?,?)", gi.Id, json.RawMessage(jsonData))
		if err != nil {
			return err
		}
		for i := range gi.Infos {
			gii := &gi.Infos[i]
			jsonData, err = json.Marshal(gii)
			if err != nil {
				return err
			}
			_, err = tx.Exec("INSERT INTO gvg.guild(group_id,guild_id,guild_data) VALUES (?,?,?)", gi.Id, gii.Id, json.RawMessage(jsonData))
			if err != nil {
				return err
			}
			for index := range gii.Data {
				bi := &gii.Data[index]
				jsonData, err = json.Marshal(bi)
				if err != nil {
					return err
				}
				_, err = tx.Exec("INSERT INTO gvg.build(group_id,guild_id,build_id,build_info) VALUES (?,?,?,?)", gi.Id, gii.Id, bi.Id, json.RawMessage(jsonData))
				if err != nil {
					return err
				}
				for roleIndex := range bi.Roles {
					role := &bi.Roles[roleIndex]
					_, err = tx.Exec("INSERT INTO gvg.reward(activity_id,role_id) VALUES (?,?)", this_.activeId, role.RoleId)
					if err != nil {
						return err
					}
				}
			}
			_, err = tx.Exec("UPDATE gvg.signup SET is_match=1 WHERE guild_id=?", gii.Id)
			if err != nil {
				return err
			}
		}
		return nil
	})
	errmsg.Must(err)
}

func (this_ *Service) startMatch() {
	timer.AfterFunc(time.Second*5, func() { // 预留30秒处理未完成的报名。但是此时客户端请求已经无效
		this_.signupQueue.PostFuncQueue(func() {
			safego.GO(func(i interface{}) {
				this_.log.Error("match panic", zap.Any("panic info", i))
				this_.match()
			}, func() {
				this_.match()
			})
		})
	})
}

// 结算
func (this_ *Service) startSettle() {
	this_.log.Info("开始结算", zap.Int64("activityId", this_.activeId))
	guildContend := rule.MustGetReader(nil).GuildContend.List()
	var guildRankCnf []*rule_model.GuildContend
	var personalRankCnf []*rule_model.GuildContend
	for i := range guildContend {
		v := &guildContend[i]
		if v.RankType == 1 {
			guildRankCnf = append(guildRankCnf, v)
		} else if v.RankType == 2 {
			personalRankCnf = append(personalRankCnf, v)
		}
	}
	sort.Slice(guildRankCnf, func(i, j int) bool {
		return guildRankCnf[i].GuildRank[0] < guildRankCnf[j].GuildRank[0]
	})
	sort.Slice(personalRankCnf, func(i, j int) bool {
		return personalRankCnf[i].GuildRank[0] < personalRankCnf[j].GuildRank[0]
	})
	getRewardFunc := func(rank int64, cnf []*rule_model.GuildContend) *rule_model.GuildContend {
		for _, v := range cnf {
			lgr := len(v.GuildRank)
			if lgr == 1 {
				if v.GuildRank[0] == rank {
					return v
				}
			} else if lgr == 2 {
				gr1 := v.GuildRank[1]
				if gr1 == 0 {
					gr1 = math.MaxInt64
				}
				if v.GuildRank[0] <= rank && gr1 >= rank {
					return v
				}
			}
		}
		return nil
	}
	c := ctx.GetContext()
	for _, group := range this_.groupMap {
		this_.calcGuildScoreGroup(group)
		SortGuildInfos(group.Infos)
		now := time.Now().UTC()

		c.StartTime = now.UnixNano()
		c.ServerId = this_.serverId
		c.ServerType = this_.serverType
		for guildIndex := range group.Infos {
			guildRank := int64(guildIndex + 1)
			guild := &group.Infos[guildIndex]
			var guildName string
			daoGuild, err := guildDao.NewGuild(guild.Id).Get(ctx.GetContext())
			if err == nil {
				guildName = daoGuild.Name
			}
			gcCnf := getRewardFunc(guildRank, guildRankCnf)
			if gcCnf != nil {
				req := &lessservicepb.Guild_UpdateBlessingEfficRequest{
					GuildId:   daoGuild.Id,
					Effic:     gcCnf.GuildBlessing,
					ExpiredAt: now.Unix() + gcCnf.GuildBlessingTime*3600,
				}
				if gcCnf.GuildBlessing > 0 && gcCnf.GuildBlessingTime > 0 {
					out := &lessservicepb.Guild_UpdateBlessingEfficResponse{}
					err := this_.nc.RequestWithOut(c, this_.gameServerId, req, out)
					if err != nil {
						this_.log.Warn("notify Guild_UpdateBlessingEfficRequest failed", zap.Error(err))
					}
				}
				roleIds := make([]string, 0, 128)
				for buildIndex := range guild.Data {
					build := &guild.Data[buildIndex]
					for roleIndex := range build.Roles {
						role := &build.Roles[roleIndex]
						roleIds = append(roleIds, role.RoleId)
					}
				}
				if len(roleIds) == 0 {
					continue
				}
				c := ctx.GetContext()
				roleMap, err := userDao.GetRoles(c, roleIds)
				if err != nil {
					this_.log.Error("userDao.GetRoles failed", zap.Error(err), zap.Strings("roles", roleIds))
				} else {
					userIds := make([]string, 0, len(roleMap))
					for _, v := range roleMap {
						userIds = append(userIds, v.UserId)
					}
					if len(userIds) == 0 {
						continue
					}
					userMap, err := userDao.GetUsers(c, userIds)
					if err != nil {
						this_.log.Error("userDao.GetRoles failed", zap.Error(err), zap.Strings("roles", roleIds))
					} else {
						reward := make([]*models.Item, 0, 2)
						for k, v := range gcCnf.Reward {
							reward = append(reward, &models.Item{ItemId: k, Count: v})
						}
						for _, v := range userMap {
							err = this_.nc.Publish(v.ServerId, &models.ServerHeader{
								StartTime:  timer.Now().UnixNano(),
								RoleId:     v.RoleId,
								ServerId:   this_.serverId,
								ServerType: this_.serverType,
								UserId:     v.UserId,
								InServerId: v.ServerId,
							}, &servicepb.Mail_SendMailFromOtherServer{Mail: &models.Mail{
								Type:       models.MailType_MailTypeSystem,
								TextId:     100032,
								Attachment: reward,
								Args:       []string{guildName, strconv.Itoa(int(guild.Score)), strconv.Itoa(int(guildRank))},
							}})
							if err != nil {
								this_.log.Error("send gvg guild reward mail error", zap.Error(err), zap.String("roleId", v.RoleId))
							} else {
								_, _, err := this_.mysql.Exec("UPDATE gvg.reward SET send_guild_reward=1 WHERE activity_id=? AND role_id=?", this_.activeId, v.RoleId)
								if err != nil {
									this_.log.Error("update gvg.reward send_guild_reward failed", zap.Error(err), zap.String("roleId", v.RoleId))
								}
							}
							// 发送个人排名奖励
							personalRank, ok := group.sl.GetRank(v.RoleId)
							if ok {
								cnf := getRewardFunc(int64(personalRank), personalRankCnf)
								personalReward := make([]*models.Item, 0, 2)
								for itemId, itemCount := range cnf.Reward {
									personalReward = append(personalReward, &models.Item{ItemId: itemId, Count: itemCount})
								}
								ugi, ok := this_.userGroupInfo[v.RoleId]
								if ok {
									err = this_.nc.Publish(v.ServerId, &models.ServerHeader{
										StartTime:  now.UnixNano(),
										RoleId:     v.RoleId,
										ServerId:   this_.serverId,
										ServerType: this_.serverType,
										UserId:     v.UserId,
										InServerId: v.ServerId,
									}, &servicepb.Mail_SendMailFromOtherServer{Mail: &models.Mail{
										Type:       models.MailType_MailTypeSystem,
										TextId:     100033,
										Attachment: personalReward,
										Args:       []string{strconv.Itoa(int(ugi.Score)), strconv.Itoa(personalRank)},
									}})
									if err != nil {
										this_.log.Error("send gvg guild reward mail error", zap.Error(err), zap.String("roleId", v.RoleId))
									} else {
										_, _, err := this_.mysql.Exec("UPDATE gvg.reward SET send_personal_reward=1 WHERE activity_id=? AND role_id=?", this_.activeId, v.RoleId)
										if err != nil {
											this_.log.Error("update gvg.reward send_guild_reward failed", zap.Error(err), zap.String("roleId", v.RoleId))
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 开始匹配的时候，需要删除上一次活动的战斗数据，并且修改表名
	err := this_.mysql.WithTxx(func(tx *sqlx.Tx) error {
		alterTime := time.Now().UTC().Format("20060102150405")
		_, err := tx.Exec("ALTER TABLE gvg.signup RENAME TO gvg.signup_" + alterTime)
		if err != nil {
			return err
		}
		_, err = tx.Exec(createTableMap["signup"])
		if err != nil {
			return err
		}

		_, err = tx.Exec("ALTER TABLE gvg.group RENAME TO gvg.group_" + alterTime)
		if err != nil {
			return err
		}
		_, err = tx.Exec(createTableMap["group"])
		if err != nil {
			return err
		}
		_, err = tx.Exec("ALTER TABLE gvg.guild RENAME TO gvg.guild_" + alterTime)
		if err != nil {
			return err
		}
		_, err = tx.Exec(createTableMap["guild"])
		if err != nil {
			return err
		}
		_, err = tx.Exec("ALTER TABLE gvg.build RENAME TO gvg.build_" + alterTime)
		if err != nil {
			return err
		}
		_, err = tx.Exec(createTableMap["build"])
		if err != nil {
			return err
		}
		_, err = tx.Exec("ALTER TABLE gvg.fighting RENAME TO gvg.fighting_" + alterTime)
		if err != nil {
			return err
		}
		_, err = tx.Exec(createTableMap["fighting"])
		if err != nil {
			return err
		}
		return nil
	})
	errmsg.Must(err)
	this_.signupMap = map[string]signupInfo{}
	this_.groupMap = map[int64]*GroupInfo{}
	this_.userGroupInfo = map[string]*BuildRole{}
	this_.guildGroupInfo = map[string]int64{}
	this_.SetGVGStatus(StatusSignup)
	ac := this_.GetConfig()
	now := time.Now().UTC()
	thisWeekBeginTime := timer.BeginOfWeek(now)
	this_.activeEndTime = thisWeekBeginTime.Add(time.Duration((ac.endDay-1)*86400+ac.closeDayTime*3600) * time.Second).Unix() // 这个星期的活动开始时间
	this_.activeStartTime = thisWeekBeginTime.Add(time.Duration((ac.startDay-1)*86400+ac.startDayTime*3600) * time.Second).Unix()
	oldActiveId := this_.activeId
	this_.activeId = ActiveIDWithTime(now)
	if oldActiveId == this_.activeId {
		thisWeekBeginTime = timer.BeginOfWeek(now.Add(time.Hour * 24 * 7))
		this_.activeEndTime = thisWeekBeginTime.Add(time.Duration((ac.endDay-1)*86400+ac.closeDayTime*3600) * time.Second).Unix() // 这个星期的活动开始时间
		this_.activeStartTime = thisWeekBeginTime.Add(time.Duration((ac.startDay-1)*86400+ac.startDayTime*3600) * time.Second).Unix()
		this_.activeId = NextActiveID(now)
	}

	this_.log.Info("结算完成。重新进入报名阶段", zap.Int64("activityId", this_.activeId))
}

type activeConfig struct {
	startDay     int64
	endDay       int64
	startDayTime int64
	closeDayTime int64
}

func (this_ *Service) GetConfig() activeConfig {
	r := rule.MustGetReader(nil)
	startDay, ok := r.KeyValue.GetInt64("GuildContendOpenDay")
	utils.MustTrue(ok)
	endDay, ok := r.KeyValue.GetInt64("GuildContendCloseDay")
	utils.MustTrue(ok)
	startDayTime, ok := r.KeyValue.GetInt64("GuildContendOpenTime")
	utils.MustTrue(ok)
	closeDayTime, ok := r.KeyValue.GetInt64("GuildContendCloseTime")
	utils.MustTrue(ok)

	return activeConfig{
		startDay:     startDay,
		endDay:       endDay,
		startDayTime: startDayTime,
		closeDayTime: closeDayTime,
	}
}

func (this_ *Service) monitoringStatus() {
	ac := this_.GetConfig()
	now := timer.Now().UTC()
	thisWeekBeginTime := timer.BeginOfWeek(now)
	weekSeconds := now.Unix() // 这个星期过了多少秒
	if this_.activeStartTime == 0 && this_.activeEndTime == 0 {
		this_.activeEndTime = thisWeekBeginTime.Add(time.Duration((ac.endDay-1)*86400+ac.closeDayTime*3600) * time.Second).Unix()     // 这个星期的活动开始时间
		this_.activeStartTime = thisWeekBeginTime.Add(time.Duration((ac.startDay-1)*86400+ac.startDayTime*3600) * time.Second).Unix() // 这个星期的活动结束时间
	}
	if weekSeconds >= this_.activeStartTime && weekSeconds < this_.activeEndTime { // 活动时间内
		// 因为是启动。所以检查是否匹配完成。如果没有匹配完成接着匹配
		this_.SetGVGStatus(StatusMatch)
		this_.startMatch()
	}
	timer.Ticker(time.Second*5, func() bool {
		this_.signupQueue.PostFuncQueue(this_.checkOnceStatus)
		return true
	})
}

func (this_ *Service) CurrWeekEndTime(now time.Time) int64 {
	ac := this_.GetConfig()
	thisWeekBeginTime := timer.BeginOfWeek(now)
	activeEndTime := thisWeekBeginTime.Add(time.Duration((ac.endDay-1)*86400+ac.closeDayTime*3600) * time.Second).Unix()
	return activeEndTime
}

func (this_ *Service) checkOnceStatus() {
	now := time.Now().UTC()
	var activeId int64
	nowSeconds := now.Unix()
	if nowSeconds > this_.CurrWeekEndTime(now) {
		activeId = NextActiveID(now)
	} else {
		activeId = ActiveIDWithTime(now)
	}
	if activeId == this_.activeId {
		nowSeconds := now.Unix() // 当前时间
		if nowSeconds >= this_.activeStartTime && nowSeconds < this_.activeEndTime {
			status := this_.GetGVGStats()
			if status != StatusMatch && status != StatusFighting {
				this_.log.Info("gvg_status: 进入匹配状态")
				this_.SetGVGStatus(StatusMatch)
				this_.startMatch()
			}
		} else if nowSeconds > this_.activeEndTime { // 进入结算状态
			if this_.GetGVGStats() != StatusSettlement {
				this_.SetGVGStatus(StatusSettlement)
				this_.startSettle()
			}
		} else {
			if this_.GetGVGStats() != StatusSignup {
				this_.SetGVGStatus(StatusSignup)
			}
		}
	} else {
		if this_.GetGVGStats() != StatusSettlement {
			this_.SetGVGStatus(StatusSettlement)
			this_.startSettle()
		}
	}
	this_.log.Info("gvg_status", zap.Int64("activeId", this_.activeId), zap.String("status", this_.GetGVGStats().String()))
}
