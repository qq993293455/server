package iggsdk

type SdkTokenCfg struct {
	BattleToken string `json:"battle_token"`
	ResToken    string `json:"res_token"`
}

type SdkGameId struct {
	AndroidEN int64 `json:"android_en"`
	IOSCN     int64 `json:"ios_cn"`
	AndroidTw int64 `json:"android_tw"`
	IOSEN     int64 `json:"ios_en"`
}
