package main

import (
	"os"
	"path"
	"strconv"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	_ "coin-server/game-server/env"
	"coin-server/rule"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

func main() {
	log := logger.MustNew(zap.DebugLevel, &logger.Options{
		Console:     "stdout",
		RemoteAddr:  nil,
		InitFields:  map[string]interface{}{},
		Development: true,
		Discard:     false,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	rule.Load(cnf)

	gens := make([]func() (string, *Dict), 0)
	gens = append(gens, GenHeroDict)
	gens = append(gens, GenItemDict)
	gens = append(gens, GenMainTaskDict)
	gens = append(gens, GenMainTaskActionDict)
	gens = append(gens, GenTitleDict)
	gens = append(gens, GenPvpEventTypeDict)

	dir := "./dictionary"
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	for _, gen := range gens {
		fname, dict := gen()
		p := path.Join(dir, fname)
		f, err := os.Create(p)
		utils.Must(err)
		data, err := jsoniter.Marshal(dict)
		utils.Must(err)
		_, err = f.Write(data)
		utils.Must(err)
	}
}

func i18n(idStr string) string {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return idStr
	}
	cfg, ok := rule.MustGetReader(nil).LanguageBackend.GetLanguageBackendById(int64(id))
	if !ok {
		logger.DefaultLogger.Warn("language not found: " + idStr)
		return idStr
	}
	return cfg.Cn
}

type Dict struct {
	Name    string         `json:"name"`
	Code    string         `json:"code"`
	DictMap map[string]any `json:"dict_map"`
}

func GenHeroDict() (filename string, dict *Dict) {
	dict = &Dict{
		Name:    "英雄名称",
		Code:    "hero_id",
		DictMap: map[string]any{},
	}
	for _, cfg := range rule.MustGetReader(nil).RowHero.List() {
		dict.DictMap[strconv.Itoa(int(cfg.Id))] = i18n(cfg.OccName) + "-" + i18n(cfg.Name)
	}
	return "hero.json", dict
}

func GenItemDict() (filename string, dict *Dict) {
	dict = &Dict{
		Name:    "物品名称",
		Code:    "item_id",
		DictMap: map[string]any{},
	}
	for _, cfg := range rule.MustGetReader(nil).Item.List() {
		dict.DictMap[strconv.Itoa(int(cfg.Id))] = i18n(cfg.Name)
	}
	return "item.json", dict
}

func GenMainTaskDict() (filename string, dict *Dict) {
	dict = &Dict{
		Name:    "主线任务",
		Code:    "main_task_id",
		DictMap: map[string]any{},
	}
	for _, cfg := range rule.MustGetReader(nil).MainTask.List() {
		dict.DictMap[strconv.Itoa(int(cfg.Id))] = i18n(cfg.Name)
	}
	return "main_task.json", dict
}

func GenMainTaskActionDict() (filename string, dict *Dict) {
	dict = &Dict{
		Name: "主线任务动作",
		Code: "main_task_action",
		DictMap: map[string]any{
			"1": "接取",
			"2": "完成",
		},
	}
	return "main_task_action.json", dict
}

func GenTitleDict() (filename string, dict *Dict) {
	dict = &Dict{
		Name:    "头衔",
		Code:    "title",
		DictMap: map[string]any{},
	}
	for _, cfg := range rule.MustGetReader(nil).RoleLvTitle.List() {
		dict.DictMap[strconv.Itoa(int(cfg.Id))] = i18n(cfg.NameId)
	}
	return "title.json", dict
}

func GenPvpEventTypeDict() (filename string, dict *Dict) {
	dict = &Dict{
		Name: "pvp事件类型",
		Code: "pvp_event_type",
		DictMap: map[string]any{
			"1": "开始",
			"2": "胜利",
			"3": "失败",
			"4": "退出",
		},
	}
	return "pvp_event_type.json", dict
}
