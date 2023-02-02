package models

const (
	CommonEventCategory = "common"
)

const (
	LaunchEventType      = "launch"
	LoginEventType       = "login"
	RegisterEventType    = "register"
	GotItemEventType     = "got_item"
	SubItemEventType     = "sub_item"
	RitualEventType      = "ritual" // 救治仪式
	GameEventType        = "game"
	BattleEventType      = "battle"
	PVPEventType         = "pvp"
	ForgeEventType       = "forge"     // 装备打造
	MainTaskEventType    = "main_task" // 主线任务
	PlayerLevelEventType = "player_level"
	PlayerTitleEventType = "player_title"
	RoguelikeEventType   = "roguelike"
	EventGamesEventType  = "event_games" // 小游戏
	ArenaEventType       = "arena"
	TowerEventType       = "tower"
	LogoutEventType      = "logout"
)

var EventTypeList = []string{
	LaunchEventType,
	LoginEventType,
	RegisterEventType,
	GotItemEventType,
	SubItemEventType,
	RitualEventType,
	GameEventType,
	BattleEventType,
	PVPEventType,
	ForgeEventType,
	MainTaskEventType,
	PlayerLevelEventType,
	PlayerTitleEventType,
	RoguelikeEventType,
	EventGamesEventType,
	ArenaEventType,
	TowerEventType,
	LogoutEventType,
}
