// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type HeroDeviationEntry struct {
	Id        values.Integer `mapstructure:"id" json:"id"`
	EntryName string         `mapstructure:"entry_name" json:"entry_name"`
}

// parse func
func ParseHeroDeviationEntry(data *Data) {
	if err := data.UnmarshalKey("hero_deviation_entry", &h.heroDeviationEntry); err != nil {
		panic(errors.New("parse table HeroDeviationEntry err:\n" + err.Error()))
	}
	for i, el := range h.heroDeviationEntry {
		h.heroDeviationEntryMap[el.Id] = i
	}
}

func (i *HeroDeviationEntry) Len() int {
	return len(h.heroDeviationEntry)
}

func (i *HeroDeviationEntry) List() []HeroDeviationEntry {
	return h.heroDeviationEntry
}

func (i *HeroDeviationEntry) GetHeroDeviationEntryById(id values.Integer) (*HeroDeviationEntry, bool) {
	index, ok := h.heroDeviationEntryMap[id]
	if !ok {
		return nil, false
	}
	return &h.heroDeviationEntry[index], true
}
