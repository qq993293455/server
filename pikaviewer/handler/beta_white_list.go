package handler

import (
	"time"

	"coin-server/common/proto/dao"
	"coin-server/pikaviewer/model"
	"coin-server/pikaviewer/utils"
)

type BetaWhiteList struct {
}

func (h *BetaWhiteList) List() ([]*dao.BetaWhiteList, error) {
	list, err := model.NewBetaWhiteList().GetAll()
	if err != nil {
		return nil, utils.NewDefaultErrorWithMsg(err.Error())
	}
	return list, nil
}

func (h *BetaWhiteList) Save(device string, enable bool, comment string) error {
	data := &dao.BetaWhiteList{
		Device:    device,
		Enable:    enable,
		Comment:   comment,
		UpdatedAt: time.Now().Unix(),
	}
	if err := model.NewBetaWhiteList().Save(data); err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}

func (h *BetaWhiteList) Del(device string) error {
	if err := model.NewBetaWhiteList().Del(device); err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}

func (h *BetaWhiteList) IsInBetaWhiteList(device string) (bool, error) {
	bw, ok, err := model.NewBetaWhiteList().GetOne(device)
	if err != nil {
		return false, utils.NewDefaultErrorWithMsg(err.Error())
	}
	if !ok {
		return false, nil
	}
	return bw.Enable, nil
}
