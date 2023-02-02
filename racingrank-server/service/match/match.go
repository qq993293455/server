package match

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/natsclient"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	readpb "coin-server/common/proto/role_state_read"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/protocol"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/racingrank-server/dao"

	"github.com/gogo/protobuf/types"

	"go.uber.org/zap"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Addr          []string `json:"addr"`
	Topic         string   `json:"topic"`
	Group         string   `json:"group"`
	ConsumerCount int      `json:"consumer_count"`
}

type Consumer struct {
	nc      *natsclient.ClusterClient
	dao     *dao.Dao
	log     *logger.Logger
	kr      *kafka.Reader
	close   chan struct{}
	closed1 int32
}

type Matcher struct {
	matcher []*Consumer
}

var matcher *Matcher

func Init(config *consulkv.Config, urls []string, serverId values.ServerId, serverType models.ServerType, log *logger.Logger) {
	nc := natsclient.NewClusterClient(serverType, serverId, urls, log)

	cfg := &Config{}
	utils.Must(config.Unmarshal("racingrank/kafka", cfg))
	list := make([]*Consumer, 0)
	for i := 0; i < cfg.ConsumerCount; i++ {
		list = append(list, NewMatcher(cfg, nc, dao.GetDao(), log))
	}
	matcher = &Matcher{list}
}

func NewMatcher(conf *Config, nc *natsclient.ClusterClient, _dao *dao.Dao, log *logger.Logger) *Consumer {
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        conf.Addr,
		GroupID:        conf.Group,
		Topic:          conf.Topic,
		MinBytes:       1,
		MaxBytes:       50e6, // 50MB
		CommitInterval: 100 * time.Millisecond,
	})
	return &Consumer{
		nc:    nc,
		dao:   _dao,
		log:   log,
		kr:    kr,
		close: make(chan struct{}),
	}
}

func Start() {
	for _, consumer := range matcher.matcher {
		go consumer.Start()
	}
}

func Close(ctx context.Context, log *logger.Logger) {
	for _, consumer := range matcher.matcher {
		if err := consumer.Close(ctx); err != nil {
			log.Error("close racing rank matcher error", zap.Error(err))
		} else {
			log.Info("close racing rank matcher success")
		}
	}
}

func GetCronExecutor() func(cron *Cron) {
	if len(matcher.matcher) <= 0 {
		panic("matcher is empty")
	}
	return matcher.matcher[0].cronExecutor
}

var protoPool = sync.Pool{New: func() interface{} { return new(pbdao.RacingRankMatch) }}

const maxCallCount = 5

func protoPoolPut(p *pbdao.RacingRankMatch) {
	p.Reset()
	protoPool.Put(p)
}

func (c *Consumer) Start() {
	ctx := context.Background()
	var proto *pbdao.RacingRankMatch

	for {
		select {
		case <-c.close:
			atomic.StoreInt32(&c.closed1, 1)
			return
		default:
		}
		// ctx, cancel := context.WithTimeout(ctx, time.Millisecond*2000)
		msg, err := c.kr.FetchMessage(ctx)
		// cancel()
		if err != nil {
			c.log.Error("call kafka.FetchMessage failed", zap.Error(err))
			time.Sleep(time.Millisecond * 5)
			continue
		}
		proto = protoPool.Get().(*pbdao.RacingRankMatch)
		err = proto.Unmarshal(msg.Value)
		if err != nil {
			c.log.Error("call proto.Unmarshal failed", zap.Error(err))
			protoPoolPut(proto)
			time.Sleep(time.Millisecond * 5)
			continue
		}
		// 匹配并将匹配结果更新至数据库
		if err := c.matching(proto); err != nil {
			c.log.Error("matching error", zap.Error(err))
			protoPoolPut(proto)
			time.Sleep(time.Millisecond * 5)
			continue
		}
		// 将结算时间写入MySQL
		// c.dao.SaveEndTime(dao.EndTime{
		// 	RoleId:  proto.RoleId,
		// 	EndTime: proto.EndTime,
		// })
		// 创建定时任务
		// ownerCron.AddCron(&Cron{
		// 	RoleId: proto.RoleId,
		// 	When:   proto.EndTime,
		// 	Exec:   c.cronExecutor,
		// })
		// 向客户端推送匹配完成
		c.matchDone(proto.RoleId)

		protoPoolPut(proto)

		if err = c.kr.CommitMessages(ctx, msg); err != nil {
			c.log.Warn("call kafka.CommitMessages failed", zap.Error(err))
			time.Sleep(time.Millisecond * 5)
		}
	}
}

func (c *Consumer) Close(ctx context.Context) error {
	close(c.close)
	c.kr.Close()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if atomic.LoadInt32(&c.closed1) == 1 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (c *Consumer) matching(self *pbdao.RacingRankMatch) error {
	count := int(self.Count)
	var err error
	ltList := make([]*dao.Data, 0)
	ltList, err = c.fetchFormMySQL(self.CombatValue, true, ltList, count, 0)
	if err != nil {
		return err
	}
	gtList := make([]*dao.Data, 0)
	gtList, err = c.fetchFormMySQL(self.CombatValue, false, gtList, count, 0)
	if err != nil {
		return err
	}
	ltCount := 9
	gtCount := 40
	minTime := time.Now().AddDate(0, 0, -7).Unix()
	list := make([]*models.RankItem, 0)
	tempMap := make(map[values.RoleId]struct{})
	// 优先取7日内登录过的玩家
	for _, item := range ltList {
		if item.LoginTime >= minTime {
			list = append(list, &models.RankItem{
				RoleId: item.RoleId,
			})
			tempMap[item.RoleId] = struct{}{}
			ltCount--
			if ltCount <= 0 {
				break
			}
		}
	}
	for _, item := range gtList {
		if item.LoginTime >= minTime {
			list = append(list, &models.RankItem{
				RoleId: item.RoleId,
			})
			tempMap[item.RoleId] = struct{}{}
			gtCount--
			if gtCount <= 0 {
				break
			}
		}
	}
	// 填补低于当前战力的数据
	if ltCount > 0 {
		for _, item := range ltList {
			if _, ok := tempMap[item.RoleId]; !ok {
				list = append(list, &models.RankItem{
					RoleId: item.RoleId,
				})
				tempMap[item.RoleId] = struct{}{}
				ltCount--
				if ltCount <= 0 {
					break
				}
			}
		}
	}
	// 填补高于当前战力的数据
	if gtCount > 0 {
		for _, item := range gtList {
			if _, ok := tempMap[item.RoleId]; !ok {
				list = append(list, &models.RankItem{
					RoleId: item.RoleId,
				})
				tempMap[item.RoleId] = struct{}{}
				gtCount--
				if gtCount <= 0 {
					break
				}
			}
		}
	}
	// 可能会存在不够填补的情况，直接不区分低于或高于的，直接填补够需要的数量
	if len(list) < count {
		tempList := make([]*dao.Data, 0)
		tempList = append(tempList, ltList...)
		tempList = append(tempList, gtList...)
		for _, item := range tempList {
			if _, ok := tempMap[item.RoleId]; !ok {
				list = append(list, &models.RankItem{
					RoleId: item.RoleId,
				})
				tempMap[item.RoleId] = struct{}{}
				if len(list) >= count {
					break
				}
			}
		}
	}
	list = append(list, self.Self)
	if err := c.dao.SaveData(&pbdao.RacingRankData{
		RoleId:       self.RoleId,
		List:         list,
		Locked:       false,
		ForceRefresh: true, // 将强制更新置为true
	}); err != nil {
		return err
	}

	return nil
}

func (c *Consumer) fetchFormMySQL(combatValue values.Integer, lt bool, list []*dao.Data, count, callCount int) ([]*dao.Data, error) {
	data, err := c.dao.Find(combatValue, lt)
	if err != nil {
		return list, err
	}
	if len(data) <= 0 {
		return list, nil
	}
	list = append(list, data...)
	callCount++
	if callCount >= maxCallCount || len(list) >= count {
		return list, nil
	}
	combatValue = data[len(data)-1].HighestPower

	time.Sleep(time.Millisecond * 100)
	return c.fetchFormMySQL(combatValue, lt, list, count, callCount)
}

func (c *Consumer) matchDone(roleId values.RoleId) {
	msg := &servicepb.RacingRank_RacingRankMatchSuccessPush{}
	if err := c.nc.Publish(0, &models.ServerHeader{}, &readpb.RoleStateROnly_PushManyToClient{
		Pcs: []*readpb.RoleStateROnly_PushToClient{{
			Roles:    []values.RoleId{roleId},
			Messages: []*types.Any{protocol.MsgToAny(msg)},
		}},
	}); err != nil {
		c.log.Error("[matchDone] publish error", zap.Error(err), zap.String("role_id", roleId))
	}
	c.log.Debug("[matchDone] success", zap.String("role_id", roleId))
}

func (c *Consumer) cronExecutor(cron *Cron) {
	data, err := c.dao.GetRacingRankData(cron.RoleId)
	if err != nil {
		c.log.Error("[GetRacingRankData] cron exec error", zap.Error(err), zap.String("role_id", cron.RoleId))
		if cron.Retry() {
			cron.RetryTimes++
			ownerCron.AddCron(cron)
		}
		return
	}
	if data.Locked || len(data.List) <= 0 {
		return
	}
	roleIds := make([]values.RoleId, 0, len(data.List))
	for _, item := range data.List {
		roleIds = append(roleIds, item.RoleId)
	}
	roleMap, err := c.dao.BatchGetRole(roleIds)
	if err != nil {
		c.log.Error("[BatchGetRole] cron exec error", zap.Error(err), zap.String("role_id", cron.RoleId))
		if cron.Retry() {
			cron.RetryTimes++
			ownerCron.AddCron(cron)
		}
		return
	}
	// 更新排行榜数据（只用更新战力即可，用于排名结算）
	for i := 0; i < len(data.List); i++ {
		roleId := data.List[i].RoleId
		if role, ok := roleMap[roleId]; ok {
			data.List[i].Power = role.Power
		}
	}
	data.Locked = true
	if err := c.dao.SaveData(data); err != nil {
		c.log.Error("[SaveData] cron exec error", zap.Error(err), zap.String("role_id", cron.RoleId))
		if cron.Retry() {
			cron.RetryTimes++
			ownerCron.AddCron(cron)
		}
		return
	}
	// 删除MySQL里的持久化定时任务
	// c.dao.DeleteEndTime(cron.RoleId)

	c.log.Debug("[CronExecutor] success", zap.String("role_id", cron.RoleId))
}
