package guild

import (
	"coin-server/common/values"

	"github.com/golang-jwt/jwt"
)

type guildSystemMsgId values.Integer

const (
	guildNameBloomKey  = "guild_name_bloom_key" // 公会名布隆过滤器的key
	guildNameBloomSize = 1e6                    // 公会名布隆过滤器的大小
)

// 锁
const (
	guildLock          = "lock:guild:"                // 公会锁+公会id
	guildMemberLock    = "lock:guild:member:"         // 公会成员锁+公会id
	guildUserLock      = "lock:guild:user:"           // 公会用户锁+玩家id
	guildUserApplyLock = "lock:guild:user:apply:"     // 玩家申请锁+玩家id
	guildBlessingEffic = "lock:guild:blessing:effic:" // 公会gvg对公会祝福的加成锁+公会id
)

// 公会日志文本id（该功能已去掉）
const (
	join             guildSystemMsgId = 1 // 	{0}加入了公会。
	leave            guildSystemMsgId = 2 //	{0}离开了公会。
	promotion        guildSystemMsgId = 3 //	{0}晋升到{1}。
	demotion         guildSystemMsgId = 4 //	{0}降级到了{1}。
	leaderChange     guildSystemMsgId = 5 //	{0}把会长职位移交到{1}。
	modifyName       guildSystemMsgId = 6 //	{0}把公会名称改成{1}。
	modifyNotice     guildSystemMsgId = 7 //	{0}把公会公告改成{1}。
	levelup          guildSystemMsgId = 8 //	恭喜，公会等级提升到Lv.{0}。
	leaderAutoChange guildSystemMsgId = 9 //	公会会长已经长时间未上线，{0}成为新的会长。
)

var jwtKey = []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")

type SystemMsg struct {
	TextId guildSystemMsgId `json:"text_id"` // 文本ID
	Args   []string         `json:"args"`    // 文本参数
}

type BuildInfo struct {
	FreeCount int64
	PayCount  int64
	Cost      map[values.ItemId]values.Integer
	ResetTime int64
}

type Invite struct {
	GuildId   values.GuildId `json:"guildId"`
	GuildName string         `json:"guildName"`
	Token     string         `json:"token"`
	Private   bool           `json:"private"` // 是否为个人邀请
	Msg       string         `json:"msg"`     // 自定义邀请信息
}

type Claims struct {
	Invite
	jwt.StandardClaims
}

func getRankRange(rank values.Integer) (values.Integer, values.Integer) {
	if rank <= 0 {
		return 1, 100
	}
	v := rank % 100

	start := rank - v + 1
	end := rank + (100 - v)
	return start, end
}
