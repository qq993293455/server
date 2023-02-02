package im

type BaseResp struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type GetTokenResp struct {
	Code   int    `json:"code"`
	Error  string `json:"error"`
	Result struct {
		Expire string `json:"expire"`
		Token  string `json:"token"`
	} `json:"result"`
}

const (
	MsgTypePrivate = iota
	MsgTypeRoom
	MsgTypeBroadcast
)

const (
	ParseTypeGeneral          = 0  // 普通消息
	ParseTypeSys              = 1  // 系统通知
	ParseTypeRedPack          = 2  // 红包🧧
	ParseTypeGuildSystemMsg   = 3  // 公会系统消息
	ParseTypeShare            = 4  // 分享（目前仅装备）
	ParseTypeGuildInvite      = 5  // 公会邀请
	ParseRoguelikeInvite      = 6  // roguelike邀请
	ParseTypeClientUpdate     = 7  // 提示客户端有更新
	ParseTypeNoticeOperator   = 8  // 运营发送公告
	ParseTypeNoticeEquip      = 9  // 打造出橙色以上装备时全服公告
	ParseTypeNoticeRelics     = 10 // 抽到橙色以上遗物时线路公告
	ParseTypeNoticeBossHall   = 11 // Boss大厅击杀公告
	_                         = 12 // 停服公告 客户端占用
	ParseTypePersonalBossHelp = 13 // 个人BOSS请求帮助
	ParseChangeNickName       = 14 // 改名
	ParseTypeGVGGuild         = 15 // 工会GVG
)

type Message struct {
	Type       int    `json:"type"`
	RoleID     string `json:"role_id"`
	RoleName   string `json:"role_name"`
	TargetID   string `json:"target_id,omitempty"`
	Content    string `json:"content"`
	ParseType  int    `json:"parse_type"`  // game server 自定义的解析类型 默认0为正常聊天消息
	Extra      string `json:"extra"`       // 扩展字段 map[string]any 类型的 json 字符串
	IsMarquee  bool   `json:"is_marquee"`  // 是否是跑马灯
	IsVolatile bool   `json:"is_volatile"` // 如果该值为true，则不在聊天显示此消息
}

type RoomRole struct {
	RoomID  string   `json:"room_id"`
	RoleIDs []string `json:"role_id"`
}

type BlackListOp struct {
	RoleID  string `json:"role_id"`
	Type    int    `json:"type"`    // 类型 1 禁言 0 解禁
	Seconds int    `json:"seconds"` // 禁言多少秒
}
