package Notice

type Enum = int64

const (
	Equip    Enum = 1001 // 打造出橙色以上装备时全服公告
	Relics   Enum = 1002 // 抽到橙色以上遗物时公告
	BossHall Enum = 1003 // boss大厅击杀Boss
)
