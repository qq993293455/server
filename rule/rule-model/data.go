package rule_model

import (
	"encoding/json"

	"github.com/spf13/viper"
)

const VersionKey = "RULE-VERSION"

type Data struct {
	Viper *viper.Viper
}

func NewData(v *viper.Viper) *Data {
	return &Data{
		Viper: v,
	}
}

func (d *Data) GetVersion() string {
	return d.Viper.GetString(VersionKey)
}

func (d *Data) UnmarshalKey(key string, to interface{}) error {
	b := []byte(d.Viper.GetString(key))
	return json.Unmarshal(b, to)
}
