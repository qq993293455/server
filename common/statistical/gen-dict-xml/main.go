package main

import (
	"encoding/xml"
	"os"
	"path"
	"strconv"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	_ "coin-server/game-server/env"
	"coin-server/rule"

	"go.uber.org/zap"
)

const header = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"

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

	gens := make([]func() Items, 0)
	gens = append(gens, GenHeroDict)
	gens = append(gens, GenItemDict)
	gens = append(gens, GenMainTaskDict)
	gens = append(gens, GenMainTaskActionDict)
	gens = append(gens, GenTitleDict)
	gens = append(gens, GenPvpEventTypeDict)

	dir := "./dict-xml"
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	dict := Dict{Items: make([]Items, 0)}
	for _, gen := range gens {
		d := gen()
		dict.Items = append(dict.Items, d)
	}
	p := path.Join(dir, "dict.xml")
	f, err := os.Create(p)
	utils.Must(err)
	defer f.Close()

	data, err := xml.MarshalIndent(dict, "", "\t")
	utils.Must(err)
	_, err = f.WriteString(xml.Header)
	utils.Must(err)
	_, err = f.Write(data)
	utils.Must(err)
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
	XMLName xml.Name `xml:"metalib"`
	Items   []Items  `xml:"items"`
}

type Items struct {
	XMLName xml.Name `xml:"items"`
	Name    string   `xml:"name,attr"`
	Desc    string   `xml:"desc,attr"`
	Item    []Item   `xml:"item"`
}

type Item struct {
	XMLName xml.Name `xml:"item"`
	Name    string   `xml:"name,attr"`
	Desc    string   `xml:"desc,attr"`
}

func GenHeroDict() (dict Items) {
	dict = Items{
		Name: "hero_id",
		Desc: "英雄名称",
		Item: make([]Item, 0),
	}
	for _, cfg := range rule.MustGetReader(nil).RowHero.List() {
		dict.Item = append(dict.Item, Item{
			Name: strconv.Itoa(int(cfg.Id)),
			Desc: i18n(cfg.OccName) + "-" + i18n(cfg.Name),
		})
	}
	return dict
}

func GenItemDict() (dict Items) {
	dict = Items{
		Name: "item_id",
		Desc: "物品名称",
		Item: make([]Item, 0),
	}
	for _, cfg := range rule.MustGetReader(nil).Item.List() {
		dict.Item = append(dict.Item, Item{
			Name: strconv.Itoa(int(cfg.Id)),
			Desc: i18n(cfg.Name),
		})
	}
	return dict
}

func GenMainTaskDict() (dict Items) {
	dict = Items{
		Name: "main_task_id",
		Desc: "主线任务",
		Item: make([]Item, 0),
	}
	for _, cfg := range rule.MustGetReader(nil).MainTask.List() {
		dict.Item = append(dict.Item, Item{
			Name: strconv.Itoa(int(cfg.Id)),
			Desc: i18n(cfg.Name),
		})
	}
	return dict
}

func GenMainTaskActionDict() (dict Items) {
	dict = Items{
		Name: "main_task_action",
		Desc: "主线任务动作",
		Item: []Item{
			{Name: "1", Desc: "接取"},
			{Name: "2", Desc: "完成"},
		},
	}
	return dict
}

func GenTitleDict() (dict Items) {
	dict = Items{
		Name: "title",
		Desc: "头衔",
		Item: make([]Item, 0),
	}
	for _, cfg := range rule.MustGetReader(nil).RoleLvTitle.List() {
		dict.Item = append(dict.Item, Item{
			Name: strconv.Itoa(int(cfg.Id)),
			Desc: i18n(cfg.NameId),
		})
	}
	return dict
}

func GenPvpEventTypeDict() (dict Items) {
	dict = Items{
		Name: "pvp_event_type",
		Desc: "pvp事件类型",
		Item: []Item{
			{Name: "1", Desc: "开始"},
			{Name: "2", Desc: "胜利"},
			{Name: "3", Desc: "失败"},
			{Name: "4", Desc: "退出"},
		},
	}
	return dict
}
