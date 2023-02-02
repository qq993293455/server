package condition

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

var parser = &CondParser{
	single:  map[models.TaskType]map[values.Integer]map[values.Integer]map[values.Integer][]values.TaskId{},
	counter: map[models.TaskType]map[values.Integer]struct{}{},
}

func GetParser() *CondParser {
	return parser
}

type CondParser struct {
	single  map[models.TaskType]map[values.Integer]map[values.Integer]map[values.Integer][]values.Integer
	counter map[models.TaskType]map[values.Integer]struct{}
}

func (c *CondParser) ParseCounter(typ, k values.Integer) {
	if c.counter[models.TaskType(typ)] == nil {
		c.counter[models.TaskType(typ)] = map[values.Integer]struct{}{}
	}
	c.counter[models.TaskType(typ)][k] = struct{}{}
}

func (c *CondParser) IsCount(typ models.TaskType, k values.Integer) bool {
	if c.counter[typ] == nil {
		return false
	}
	_, ok := c.counter[typ][k]
	return ok
}

func (c *CondParser) ParseSingle(typ, k, v, system, id values.Integer) {
	if c.single[models.TaskType(typ)] == nil {
		c.single[models.TaskType(typ)] = map[values.Integer]map[values.Integer]map[values.Integer][]values.Integer{}
	}
	if c.single[models.TaskType(typ)][k] == nil {
		c.single[models.TaskType(typ)][k] = map[values.Integer]map[values.Integer][]values.Integer{}
	}
	if c.single[models.TaskType(typ)][k][v] == nil {
		c.single[models.TaskType(typ)][k][v] = map[values.Integer][]values.Integer{}
	}
	if c.single[models.TaskType(typ)][k][v][system] == nil {
		c.single[models.TaskType(typ)][k][v][system] = make([]values.Integer, 0)
	}
	c.single[models.TaskType(typ)][k][v][system] = append(c.single[models.TaskType(typ)][k][v][system], id)
}
