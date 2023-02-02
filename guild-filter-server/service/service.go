package service

import (
	"context"
	"time"

	ctx2 "coin-server/common/ctx"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/orm"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/service"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/guild-filter-server/service/guild"
	"coin-server/rule"

	"github.com/jinzhu/now"

	"go.uber.org/zap"
)

type Service struct {
	svc        *service.Service
	log        *logger.Logger
	serverId   values.ServerId
	serverType models.ServerType
	guildData  map[values.GuildId]guild.Guild
}

func NewService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
) *Service {
	svc := service.NewService(urls, log, serverId, serverType, true, false, eventlocal.CreateEventLocal(true))
	s := &Service{
		svc:        svc,
		log:        log,
		serverId:   serverId,
		serverType: serverType,
		guildData:  make(map[values.GuildId]guild.Guild),
	}
	return s
}

func (svc *Service) Serve() {
	svc.loadGuildData()
	svc.Router()
	svc.svc.Start(func(event interface{}) {
		svc.log.Warn("unknown event", zap.Any("event", event))
	}, true)
}

func (svc *Service) Stop() {
	svc.svc.Close()
}

func (svc *Service) Router() {
	guild.NewGuildService(svc.serverId, svc.serverType, svc.svc, svc.guildData, svc.log).Router()
}

const page = 200 // 每次查询的数量

func (svc *Service) loadGuildData() {
	c := redisclient.GetGuildRedis()
	ctx := context.Background()
	expected := 0
	tempMap := make(map[values.GuildId]struct{})
	sevenDaysKey := get7DaysKey()
	for _, key := range enum.GuildAllIdKey {
		index := uint64(0)
		for {
			var keys []string
			var err error
			keys, index, err = c.HScan(ctx, key, index, "*", page).Result()
			if err != nil {
				panic(err)
			}
			newKeyS := make([]orm.RedisInterface, 0, len(keys)/2)
			for i := 0; i < len(keys); i += 2 {
				m := &pbdao.Guild{
					Id: keys[i],
				}
				newKeyS = append(newKeyS, m)
				expected++
			}

			notFound, e := orm.GetOrm(ctx).MGetPB(c, newKeyS...)
			if e != nil {
				panic(e)
			}
			notFoundMap := make(map[int]bool, len(notFound))
			for _, v := range notFound {
				notFoundMap[v] = true
			}

			for i := 0; i < len(newKeyS); i++ {
				if _, ok := notFoundMap[i]; ok {
					continue
				}
				guildDao := newKeyS[i].(*pbdao.Guild)
				temp := guild.NewGuild(guildDao)
				// HScan 可能会返回重复的key
				if _, ok := tempMap[temp.Id]; ok {
					continue
				}
				memberKey := utils.GenDefaultRedisKey(values.GuildMember, values.Hash, temp.Id)
				members := make([]*pbdao.GuildMember, 0)
				if err := orm.GetOrm(ctx).HGetAll(c, memberKey, &members); err != nil {
					panic(err)
				}
				temp.Full = values.Integer(len(members)) >= getMaxMemberCount(guildDao.Level)
				tempMap[temp.Id] = struct{}{}
				temp.Active = getAllMember7DayActive(members, sevenDaysKey)
				svc.guildData[temp.Id] = temp
			}
			if index == 0 || len(keys) != page {
				break
			}
		}
	}
	d := expected - len(svc.guildData)
	if d > 0 {
		svc.log.Warn("load guild data finish", zap.Int("expected", expected), zap.Int("actual", len(svc.guildData)), zap.Int("diff", d))
	} else {
		svc.log.Info("load guild data finish", zap.Int("expected", expected), zap.Int("actual", len(svc.guildData)))
	}
}

func get7DaysKey() []values.Integer {
	list := make([]values.Integer, 0, 7)
	today := now.BeginningOfDay()
	list = append(list, today.Unix())
	for i := 1; i < 7; i++ {
		today = today.Add(-24 * time.Hour * time.Duration(i))
		list = append(list, today.Unix())
	}
	return list
}

func get7DayActive(member *pbdao.GuildMember, sevenDaysKey []values.Integer) values.Integer {
	var total values.Integer
	for _, key := range sevenDaysKey {
		total += member.ActiveValue[key]
	}
	return total
}

func getAllMember7DayActive(member []*pbdao.GuildMember, sevenDaysKey []values.Integer) values.Integer {
	var total values.Integer
	for _, v := range member {
		total += get7DayActive(v, sevenDaysKey)
	}
	return total
}

func getMaxMemberCount(level values.Level) values.Integer {
	item, ok := rule.MustGetReader(ctx2.GetContext()).Guild.GetGuildById(level)
	if !ok {
		return 0
	}
	var count values.Integer
	for _, v := range item.MembersCount {
		count += v
	}
	return count
}
