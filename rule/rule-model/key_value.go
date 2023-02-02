package rule_model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"coin-server/common/proto/models"
	"coin-server/common/values"
)

const (
	itemTyp       = "item"
	integerArray  = "repeated int64"
	float64Array  = "repeated float64"
	itemArray     = "repeated item"
	mapInt64Int64 = "map<int64,int64>"
)

var list = []string{"", ""}

type KeyValue struct {
	Key   string
	Value interface{}
}

func ParseKeyValue(data *Data) {
	list := make([]map[string]string, 0)
	if err := data.UnmarshalKey("key_value", &list); err != nil {
		panic(errors.New("parse table KeyValue err:\n" + err.Error()))
	}
	for _, el := range list {
		k := el["key"]
		v := el["value"]
		kv := KeyValue{k, v}
		typ := strings.ToLower(el["type"])
		if typ == itemArray {
			items := strings.Split(v, ",")
			kv.Value = []*models.Item{}
			for _, itemstr := range items {
				item := strings.Split(itemstr, ":")
				if len(item) != 2 {
					continue
				}
				id, _ := strconv.Atoi(item[0])
				num, _ := strconv.Atoi(item[1])

				kv.Value = append(kv.Value.([]*models.Item), &models.Item{
					ItemId: int64(id),
					Count:  int64(num),
				})
			}
		} else if typ == itemTyp {
			temp := strings.Split(v, ":")
			id, _ := strconv.Atoi(temp[0])
			num, _ := strconv.Atoi(temp[1])
			kv.Value = models.Item{
				ItemId: int64(id),
				Count:  int64(num),
			}
		} else if typ == integerArray {
			temp := strings.Split(v, ",")
			array := make([]values.Integer, len(temp))
			for idx := range temp {
				id, err := strconv.Atoi(temp[idx])
				if err != nil {
					panic(err)
				}
				array[idx] = int64(id)
			}
			kv.Value = array
		} else if typ == float64Array {
			temp := strings.Split(v, ",")
			array := make([]float64, len(temp))
			for idx := range temp {
				id, err := strconv.ParseFloat(temp[idx], 64)
				if err != nil {
					panic(err)
				}
				array[idx] = id
			}
			kv.Value = array
		} else if typ == mapInt64Int64 && k != "RenameConsumeItem" && k != "DefaultTowerCost" {
			mapData := make(map[int64]int64)
			temp := strings.Split(v, ",")
			if len(temp) <= 0 {
				panic(fmt.Errorf("invalid value form key: %s", k))
			}
			for _, item := range temp {
				s := strings.Split(item, ":")
				if len(s) != 2 {
					panic(fmt.Errorf("invalid value form key: %s", k))
				}
				mapK, err := strconv.Atoi(s[0])
				if err != nil {
					panic(fmt.Errorf("invalid value form key: %s", k))
				}
				mapV, err := strconv.Atoi(s[1])
				if err != nil {
					panic(fmt.Errorf("invalid value form key: %s", k))
				}
				mapData[int64(mapK)] = int64(mapV)
			}
			kv.Value = mapData
		}

		h.keyValue = append(h.keyValue, kv)
	}
}

func (kv *KeyValue) GetString(key string) (string, bool) {
	for _, el := range h.keyValue {
		if el.Key == key {
			if v, ok := el.Value.(string); ok {
				return v, true
			}
			return "", false
		}
	}
	return "", false
}

func (kv *KeyValue) GetInt(key string) (int, bool) {
	for _, el := range h.keyValue {
		if el.Key == key {
			temp, ok := el.Value.(string)
			if !ok {
				return 0, false
			}
			v, err := strconv.Atoi(temp)
			if err != nil {
				return 0, false
			}
			return v, true
		}
	}
	return 0, false
}

func (kv *KeyValue) GetInt64(key string) (int64, bool) {
	for _, el := range h.keyValue {
		if el.Key == key {
			temp, ok := el.Value.(string)
			if !ok {
				return 0, false
			}
			v, err := strconv.Atoi(temp)
			if err != nil {
				return 0, false
			}
			return int64(v), true
		}
	}
	return 0, false
}

func (kv *KeyValue) GetBool(key string) (bool, bool) {
	for _, el := range h.keyValue {
		if el.Key == key {
			temp, ok := el.Value.(string)
			if !ok {
				return false, false
			}
			if strings.ToLower(temp) == "true" {
				return true, true
			}
			return false, false
		}
	}
	return false, false
}

func (kv *KeyValue) GetItem(key string) (*models.Item, bool) {
	for _, el := range h.keyValue {
		if el.Key == key {
			v, ok := el.Value.(models.Item)
			if !ok {
				return nil, false
			}
			return &v, true
		}
	}
	return nil, false
}

func (kv *KeyValue) GetItemArray(key string) ([]*models.Item, bool) {
	for _, el := range h.keyValue {
		if el.Key == key {
			v, ok := el.Value.([]*models.Item)
			if !ok {
				return nil, false
			}
			return v, true
		}
	}
	return nil, false
}

func (kv *KeyValue) GetIntegerArray(key string) ([]values.Integer, bool) {
	for _, el := range h.keyValue {
		if el.Key == key {
			v, ok := el.Value.([]values.Integer)
			if !ok {
				return nil, false
			}
			return v, true
		}
	}
	return nil, false
}

func (kv *KeyValue) GetFloatArray(key string) ([]float64, bool) {
	for _, el := range h.keyValue {
		if el.Key == key {
			v, ok := el.Value.([]float64)
			if !ok {
				return nil, false
			}
			return v, true
		}
	}
	return nil, false
}

func (kv *KeyValue) GetMapInt64Int64(key string) (map[values.Integer]values.Integer, bool) {
	for _, el := range h.keyValue {
		if el.Key == key {
			v, ok := el.Value.(map[values.Integer]values.Integer)
			if !ok {
				return nil, false
			}
			return v, true
		}
	}
	return nil, false
}
