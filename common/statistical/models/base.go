package models

//var (
//	TopicModelMap = map[string]Model{}
//)

type Model interface {
	HashKey() []byte
	ToJson() []byte
	GetEventType() string
	Preset() // 预设值
}

//func init() {
//	registerModel(LoginTopic, &Login{})
//	registerModel(RegisterTopic, &Register{})
//	registerModel(GotItemTopic, &GotItem{})
//	registerModel(SubItemTopic, &SubItem{})
//	registerModel(RitualTopic, &Ritual{})
//	registerModel(GameTopic, &Game{})
//	registerModel(BattleTopic, &Battle{})
//	registerModel(PVPTopic, &Pvp{})
//	registerModel(ForgeTopic, &Forge{})
//	registerModel(MainTaskTopic, &MainTask{})
//	registerModel(EventGamesTopic, &EventGames{})
//	registerModel(PlayerLevelTopic, &PlayerLevel{})
//	registerModel(PlayerTitleTopic, &PlayerTitle{})
//	registerModel(RoguelikeTopic, &Roguelike{})
//	registerModel(ArenaTopic, &Arena{})
//	registerModel(TowerTopic, &Tower{})
//
//	for _, topic := range TopicList {
//		if _, ok := TopicModelMap[topic]; !ok {
//			panic(fmt.Errorf("topic %s need register model", topic))
//		}
//	}
//}
//
//func registerModel(topic string, model Model) {
//	if _, ok := TopicModelMap[topic]; ok {
//		panic(fmt.Errorf("topic %s has been registered", topic))
//	}
//	TopicModelMap[topic] = model
//}
