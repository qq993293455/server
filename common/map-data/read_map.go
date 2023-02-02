package map_data

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"

	"coin-server/common/ctx"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

//go:embed data
var dataDir embed.FS
var dirName = "data"
var mapDataMap = map[string]*MapData{}
var defaultMapKey = "DefaultBattleMapID"

func init() {
	var mapPaths []string
	err := fs.WalkDir(dataDir, dirName, func(path string, d fs.DirEntry, err error) error {
		if path == dirName {
			return nil
		}
		if filepath.Ext(path) == ".json" {
			mapPaths = append(mapPaths, path)
		}
		return nil
	})
	utils.Must(err)

	for _, v := range mapPaths {
		data, err := dataDir.ReadFile(v)
		utils.Must(err)
		mm := &MapData{}
		utils.Must(json.Unmarshal(data, mm))
		mapDataMap[mm.Brief.Name] = mm
	}
}

//type MonsterPos struct {
//	ConfigId     int64   `json:"_ConfigId"`
//	X            float64 `json:"_X"`
//	Y            float64 `json:"_Y"`
//	FlushSeconds int64   `json:"_FlushSeconds"`
//}
//
//type MonsterRand struct {
//	ConfigId     int64 `json:"_ConfigId"`
//	Count        int64 `json:"_Count"`
//	FlushSeconds int64 `json:"_FlushSeconds"`
//}
//
//type IndexInfo struct {
//	Index    int64 `json:"_Idx"`
//	MapState int   `json:"_MapState"`
//}

//type MapData struct {
//	Id              string         `json:"_Id"`
//	SizeX           float64        `json:"_SizeX"`
//	SizeY           float64        `json:"_SizeY"`
//	MapData         []IndexInfo    `json:"_GridDatas"`
//	AutoFlush       bool           `json:"_AutoFlush"`
//	PlayerRevive    int64          `json:"_PlayerRevive"` // 玩家是否自动复活,小于0，不复活，0 立马复活。 大于0 就是多少秒之后复活
//	Monsters        []*MonsterPos  `json:"_Monsters"`
//	MonsterRand     []*MonsterRand `json:"_MonsterRand"`
//	AutoCloseMap    bool           `json:"_AutoCloseMap"`    // 当玩家或者怪死完之后是否自动退掉地图
//	MonsterAutoMove bool           `json:"_MonsterAutoMove"` // 怪是否自动移动闲逛
//	PlayerAutoMove  bool           `json:"_PlayerAutoMove"`  // 当没有控制时，玩家是否自动移动闲逛
//}

func MustGetMapData(name string) *MapData {
	v, ok := mapDataMap[name]
	if !ok {
		panic("not found : " + name)
	}
	return v
}

func GetLogicMapId(ctx *ctx.Context, scene int64) string {
	r := rule.MustGetReader(ctx)
	s, ok := r.MapScene.GetMapSceneById(scene)
	if !ok {
		panic(fmt.Sprintf("Map Scene: %d not exist! ", scene))
	}
	return s.CfgMap
}

func GetLogicMapCnf(ctx *ctx.Context, scene int64) *rulemodel.MapScene {
	r := rule.MustGetReader(ctx)
	s, ok := r.MapScene.GetMapSceneById(scene)
	if !ok {
		panic(fmt.Sprintf("Map Scene: %d not exist! ", scene))
	}
	return s
}

func GetDefaultMapId(ctx *ctx.Context) values.MapId {
	r := rule.MustGetReader(ctx)
	mapId, ok := r.KeyValue.GetInt64(defaultMapKey)
	if !ok {
		panic(fmt.Sprintf("DefaultBattleMapID Key not found"))
	}
	return mapId
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Event struct {
	Id int64 `json:"event_id"`
	X  int   `json:"x"`
	Y  int   `json:"y"`
}

type MapData struct {
	Brief struct {
		Name   string `json:"name"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"brief"`
	MonsterGroups []interface{} `json:"monster_groups"`
	Monsters      []struct {
		MonsterId int64 `json:"monster_id"`
		X         int   `json:"x"`
		Y         int   `json:"y"`
	} `json:"monsters"`
	BirthGroups   []interface{} `json:"birth_groups"`
	Births        []Point       `json:"births"`
	RevivalGroups []interface{} `json:"revival_groups"`
	Revivals      []Point       `json:"revivals"`
	BorderGroups  []interface{} `json:"border_groups"`
	Borders       []Point       `json:"borders"`
	WallGroups    []interface{} `json:"wall_groups"`
	Walls         []Point       `json:"walls"`
	RoadGroups    []interface{} `json:"road_groups"`
	Roads         []interface{} `json:"roads"`
	CollectGroups []interface{} `json:"collect_groups"`
	Collects      []Point       `json:"collects"`
	EventGroups   []interface{} `json:"event_groups"`
	Events        []Event       `json:"events"`
}
