package personalboss

const helpLock = "lock:personal_boss_help:" // 公会锁+role_id

type HelpShare struct {
	RoleId    string `json:"role_id"`
	HelpMsgId string `json:"help_msg_id"`
}
