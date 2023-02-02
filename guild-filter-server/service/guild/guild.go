package guild

import (
	"strings"
	"sync"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/proto/guild_filter_service"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/values"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	guildData  map[values.GuildId]Guild
	lock       *sync.RWMutex
}

func NewGuildService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	guildData map[values.GuildId]Guild,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		guildData:  guildData,
		log:        log,
		lock:       &sync.RWMutex{},
	}
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("gvg信息查询", svc.GVGInfo, handler.GVGServerAuth)
	h := svc.svc.Group(handler.GameServerAuth)
	h.RegisterFunc("查找公会", svc.Find)
	h.RegisterFunc("获取可一键加入的公会", svc.CanJoinGuild)
	h.RegisterFunc("添加", svc.Add)
	h.RegisterFunc("删除", svc.Delete)

}

func (svc *Service) Find(_ *ctx.Context, req *guild_filter_service.Guild_GuildFilterFindRequest) (*guild_filter_service.Guild_GuildFilterFindResponse, *errmsg.ErrMsg) {
	svc.lock.RLock()
	defer svc.lock.RUnlock()

	limit := 10
	list := make([]string, 0)
	// 玩家通过名字查找，不需要判断战斗力
	if req.Name != "" {
		for _, guild := range svc.guildData {
			if strings.Contains(guild.Name, req.Name) && // 名字满足
				guild.Lang == guild.Lang && // 语言满足
				(req.NoAudit && guild.AutoJoin || !req.NoAudit) { // 不需要审核 || 不需要审核和审核都满足
				list = append(list, guild.Id)
				if len(list) >= limit {
					break
				}
			}
		}
	} else {
		// 推荐公会，需要判断战斗力
		for _, guild := range svc.guildData {
			if !guild.Full && // 不满员
				guild.Active > 0 && // 活跃>0
				(guild.CombatValueLimit == 0 || guild.CombatValueLimit <= req.CombatValue) && // 战斗力满足
				guild.Lang == guild.Lang && // 语言满足
				(req.NoAudit && guild.AutoJoin || !req.NoAudit) { // 不需要审核 || 不需要审核和审核都满足
				list = append(list, guild.Id)
				if len(list) >= limit {
					break
				}
			}
		}
		// 一个满足条件的都没有，去掉活跃>0的条件限制，然后再次查找
		if len(list) == 0 {
			for _, guild := range svc.guildData {
				if !guild.Full && // 不满员
					(guild.CombatValueLimit == 0 || guild.CombatValueLimit <= req.CombatValue) && // 战斗力满足
					guild.Lang == guild.Lang && // 语言满足
					(req.NoAudit && guild.AutoJoin || !req.NoAudit) { // 不需要审核 || 不需要审核和审核都满足
					list = append(list, guild.Id)
					if len(list) >= limit {
						break
					}
				}
			}
		}
	}

	return &guild_filter_service.Guild_GuildFilterFindResponse{
		Id: list,
	}, nil
}

func (svc *Service) CanJoinGuild(_ *ctx.Context, req *guild_filter_service.Guild_GuildFilterCanJoinGuildRequest) (*guild_filter_service.Guild_GuildFilterCanJoinGuildResponse, *errmsg.ErrMsg) {
	svc.lock.RLock()
	defer svc.lock.RUnlock()
	// 点击一键申请后，筛选出公会池中所有满足①无需审核②同语种③近7日活跃度不为0的公会，并加入筛选出的第1个
	// 若不存在这样的公会，则去掉同语种的条件
	// 若还不存在这样的公会，则飘出提示"暂无可直接加入的公会"
	var find string
	for _, guild := range svc.guildData {
		if !guild.Full && // 不满员
			guild.Active > 0 && // 活跃>0
			(guild.CombatValueLimit == 0 || guild.CombatValueLimit <= req.CombatValue) && // 战斗力满足
			guild.Lang == guild.Lang && // 语言满足
			guild.AutoJoin { // 不需要审核
			find = guild.Id
			break
		}
	}
	if find == "" {
		for _, guild := range svc.guildData {
			if !guild.Full && // 不满员
				(guild.CombatValueLimit == 0 || guild.CombatValueLimit <= req.CombatValue) && // 战斗力满足
				guild.AutoJoin { // 不需要审核
				find = guild.Id
				break
			}
		}
	}
	return &guild_filter_service.Guild_GuildFilterCanJoinGuildResponse{
		Id: find,
	}, nil
}

func (svc *Service) Add(_ *ctx.Context, req *guild_filter_service.Guild_GuildFilterUpdateRequest) (*guild_filter_service.Guild_GuildFilterUpdateResponse, *errmsg.ErrMsg) {
	svc.lock.Lock()
	defer svc.lock.Unlock()
	svc.guildData[req.Id] = Guild{
		Id:               req.Id,
		Name:             req.Name,
		Level:            req.Level,
		Lang:             req.Lang,
		CombatValueLimit: req.CombatValueLimit,
		AutoJoin:         req.AutoJoin,
		Active:           req.Active,
		Full:             req.Full,
		Count:            req.Count,
	}
	return &guild_filter_service.Guild_GuildFilterUpdateResponse{}, nil
}

func (svc *Service) Delete(_ *ctx.Context, req *guild_filter_service.Guild_GuildFilterDeleteRequest) (*guild_filter_service.Guild_GuildFilterDeleteResponse, *errmsg.ErrMsg) {
	svc.lock.Lock()
	defer svc.lock.Unlock()
	delete(svc.guildData, req.Id)
	return &guild_filter_service.Guild_GuildFilterDeleteResponse{}, nil
}

func (svc *Service) GVGInfo(_ *ctx.Context, req *guild_filter_service.Guild_GuildGVGInfoRequest) (*guild_filter_service.Guild_GuildGVGInfoResponse, *errmsg.ErrMsg) {
	svc.lock.RLock()
	defer svc.lock.RUnlock()
	guild, ok := svc.guildData[req.Id]
	if !ok {
		return nil, errmsg.NewErrGuildNotExist()
	}
	return &guild_filter_service.Guild_GuildGVGInfoResponse{
		Level:  guild.Level,
		Active: guild.Active,
		Count:  guild.Count,
	}, nil
}
