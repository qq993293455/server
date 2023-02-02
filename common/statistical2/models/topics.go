package models

const (
	LaunchTopic      = "row_launch"
	LoginTopic       = "row_login"
	RegisterTopic    = "row_register"
	GotItemTopic     = "row_got_item"
	SubItemTopic     = "row_sub_item"
	RitualTopic      = "row_ritual" // 救治仪式
	GameTopic        = "row_game"
	BattleTopic      = "row_battle"
	PVPTopic         = "row_pvp"
	ForgeTopic       = "row_forge"     // 装备打造
	MainTaskTopic    = "row_main_task" // 主线任务
	PlayerLevelTopic = "row_player_level"
	PlayerTitleTopic = "row_player_title"
	RoguelikeTopic   = "row_roguelike"
	EventGamesTopic  = "row_event_games" // 小游戏
	ArenaTopic       = "row_arena"
	TowerTopic       = "row_tower"
	LogoutTopic      = "row_logout"
	PayTopic         = "row_pay"
)

var TopicList = []string{
	LaunchTopic,
	LoginTopic,
	RegisterTopic,
	// GotItemTopic,
	// SubItemTopic,
	// RitualTopic,
	GameTopic,
	// BattleTopic,
	// PVPTopic,
	// ForgeTopic,
	// MainTaskTopic,
	PlayerLevelTopic,
	// PlayerTitleTopic,
	// RoguelikeTopic,
	// EventGamesTopic,
	// ArenaTopic,
	// TowerTopic,
	LogoutTopic,
	PayTopic,
}

var LoginTopicList = []string{
	LoginTopic,
	LogoutTopic,
}
