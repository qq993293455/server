package handler

import (
	"math/rand"
	"strconv"
	"sync"
	"time"

	"coin-server/pikaviewer/model"
	"coin-server/pikaviewer/utils"
)

var l sync.Mutex

const letterBytes = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

func getBatchId() (int64, error) {
	l.Lock()
	defer l.Unlock()
	dao, err := model.GetBatchPB()
	if err != nil {
		return 0, err
	}
	dao.BatchId += 1
	if err = model.SaveBatchPB(dao); err != nil {
		return 0, err
	}
	return dao.BatchId, nil
}

const (
	StableKey = iota + 1 // 固定类型
	RandomKey            // 随机类型
)

type CdKeyGen struct {
	Id       string `json:"id"  binding:"required"`
	BeginAt  int64  `db:"begin_at" json:"begin_at"  binding:"required"`
	EndAt    int64  `db:"end_at" json:"end_at"  binding:"required"`
	LimitCnt int64  `db:"limit_cnt" json:"limit_cnt"  binding:"required"`
	LimitTyp int64  `db:"limit_typ" json:"limit_typ"  binding:"required"`
	Reward   string `db:"reward" json:"reward"  binding:"required"`
	IsActive bool   `db:"is_active" json:"is_active"`
	Typ      int    `db:"typ" json:"typ" binding:"required"`
	BatchCnt int    `db:"batch_cnt" json:"batch_cnt" binding:"required"`
}

func (h *CdKeyGen) Save(req *CdKeyGen) error {
	c := model.NewCdKeyGen()
	currBatchId, err := getBatchId()
	if err != nil {
		return utils.NewDefaultErrorWithMsg("保存失败：" + err.Error())
	}
	if len(req.Reward) < 2 || req.Reward[0] != '{' || req.Reward[len(req.Reward)-1] != '}' {
		return utils.NewDefaultErrorWithMsg("奖励格式错误：" + err.Error())
	}
	if req.Typ == StableKey {
		// 固定
		exist, err := c.KeyExist(req.Id)
		if err != nil {
			return utils.NewDefaultErrorWithMsg("查询失败：" + err.Error())
		}
		if exist {
			return utils.NewDefaultErrorWithMsg("已存在该兑换码")
		}
		if err = c.Save(&model.CdKeyGen{
			Id:        req.Id,
			BatchId:   currBatchId,
			BeginAt:   req.BeginAt,
			EndAt:     req.EndAt,
			LimitCnt:  req.LimitCnt,
			LimitTyp:  req.LimitTyp,
			Reward:    req.Reward,
			IsActive:  true,
			CreatedAt: time.Now().Unix(),
		}); err != nil {
			return utils.NewDefaultErrorWithMsg("保存失败：" + err.Error())
		}
	} else {
		data := make([]*model.CdKeyGen, 0, req.BatchCnt)
		for i := 0; i < req.BatchCnt; i++ {
			data = append(data, &model.CdKeyGen{
				Id:        genUniqueKey(currBatchId),
				BatchId:   currBatchId,
				BeginAt:   req.BeginAt,
				EndAt:     req.EndAt,
				LimitCnt:  req.LimitCnt,
				LimitTyp:  req.LimitTyp,
				Reward:    req.Reward,
				IsActive:  true,
				CreatedAt: time.Now().Unix(),
			})
		}
		if err = c.Saves(data); err != nil {
			return utils.NewDefaultErrorWithMsg("保存失败：" + err.Error())
		}
	}
	return nil
}

func (h *CdKeyGen) DeActive(batchId int64) error {
	c := model.NewCdKeyGen()
	return c.DeActive(batchId)
}

func genUniqueKey(batchId int64) string {
	ret := strconv.FormatInt(batchId, 32)
	retB := []byte(ret)
	b := make([]byte, 12-len(retB), 12)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	for i := range retB {
		if retB[i] == 'o' {
			retB[i] = 'Y'
		}
		if retB[i] == 'i' {
			retB[i] = 'Z'
		}
		if retB[i] >= 'a' && ret[i] <= 'z' {
			retB[i] -= 32
		}
		b = append(b, retB[i])
	}
	for idx := 0; idx < 6; idx++ {
		tar := 6 + rand.Intn(6)
		b[idx], b[tar] = b[tar], b[idx]
	}
	return string(b)
}
