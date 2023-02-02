package serverids

import (
	"strconv"

	"coin-server/common/consulkv"
	"coin-server/common/utils"
	wr "coin-server/common/utils/weightedrand"
	"coin-server/common/values"
)

type serverIds struct {
	chooser *wr.Chooser[values.ServerId, int]
}

var sids *serverIds

func Init(cnf *consulkv.Config) {
	ids := make(map[string]int)
	err := cnf.Unmarshal("server_ids", &ids)
	utils.Must(err)

	choices := make([]*wr.Choice[values.ServerId, int], 0, len(ids))
	for sid, weight := range ids {
		id, err := strconv.Atoi(sid)
		utils.Must(err)
		choices = append(choices, wr.NewChoice(values.ServerId(id), weight))
	}

	ch, err := wr.NewChooser(choices...)
	utils.Must(err)
	sids = &serverIds{chooser: ch}
}

func (s *serverIds) Assign() values.ServerId {
	return s.chooser.Pick()
}

func Assign() values.ServerId {
	return sids.Assign()
}
