package handler

import (
	"time"

	"coin-server/common/proto/dao"
	"coin-server/pikaviewer/model"
	"coin-server/pikaviewer/utils"
)

type WhiteList struct {
}

func (h *WhiteList) List() ([]*dao.WhiteList, error) {
	list, err := model.NewWhiteList().GetAll()
	if err != nil {
		return nil, utils.NewDefaultErrorWithMsg(err.Error())
	}
	return list, nil
}

func (h *WhiteList) Save(device string, enable bool, comment string) error {
	data := &dao.WhiteList{
		Device:    device,
		Enable:    enable,
		Comment:   comment,
		UpdatedAt: time.Now().Unix(),
	}
	if err := model.NewWhiteList().Save(data); err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}

func (h *WhiteList) Del(device string) error {
	if err := model.NewWhiteList().Del(device); err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}

func (h *WhiteList) IsInWhiteList(device string) (bool, error) {
	w, ok, err := model.NewWhiteList().GetOne(device)
	if err != nil {
		return false, utils.NewDefaultErrorWithMsg(err.Error())
	}
	if !ok {
		return false, nil
	}
	return w.Enable, nil
}
