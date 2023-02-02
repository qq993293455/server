package handler

import (
	"encoding/json"
	"time"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/pikaviewer/model"
	"coin-server/pikaviewer/utils"
)

type Announcement struct {
	Id        string `json:"id"`
	Type      string `db:"type" json:"type" binding:"required"`
	Version   string `db:"version" json:"version" binding:"required"`
	BeginTime int64  `db:"begin_time" json:"begin_time" binding:"required"`
	EndTime   int64  `db:"end_time" json:"end_time"`
	StoreUrl  string `db:"store_url" json:"store_url"`
	Content   string `db:"content" json:"content" binding:"required"`
	ShowLogin bool   `db:"show_login" json:"show_login"`
}

func (h *Announcement) Find() ([]*model.Announcement, error) {
	return model.NewAnnouncement().Find()
}

func (h *Announcement) Save(req *Announcement) error {
	a := model.NewAnnouncement()
	id := req.Id
	if id == "" {
		data, err := model.NewAnnouncement().GetPB(id)
		if err != nil {
			return err
		}
		if data != nil {
			return utils.NewDefaultErrorWithMsg("当前版本下已存在同类型的公告")
		}
		id = h.getId(req.Type, req.Version)
	}
	content := make(map[int64]string)
	if err := json.Unmarshal([]byte(req.Content), &content); err != nil {
		return utils.NewDefaultErrorWithMsg("公告内容数据有误：" + err.Error())
	}
	if err := a.SavePB(&dao.Announcement{
		Id:        id,
		BeginTime: req.BeginTime,
		EndTime:   req.EndTime,
		StoreUrl:  req.StoreUrl,
		Content:   content,
		ShowLogin: req.ShowLogin,
	}); err != nil {
		return utils.NewDefaultErrorWithMsg("保存失败：" + err.Error())
	}
	if err := a.Save(&model.Announcement{
		Id:        id,
		Type:      req.Type,
		Version:   req.Version,
		BeginTime: req.BeginTime,
		EndTime:   req.EndTime,
		StoreUrl:  req.StoreUrl,
		Content:   req.Content,
		ShowLogin: req.ShowLogin,
		CreatedAt: time.Now().Unix(),
	}); err != nil {
		return utils.NewDefaultErrorWithMsg("保存失败：" + err.Error())
	}
	return nil
}

func (h *Announcement) GetPB(typ, version string, language int64, gm bool) (*models.Announcement, error) {
	data, err := model.NewAnnouncement().GetPB(h.getId(typ, version))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var content string
	if gm {
		b, err := json.Marshal(data.Content)
		if err != nil {
			return nil, utils.NewDefaultErrorWithMsg("获取公告详情失败：" + err.Error())
		}
		content = string(b)
	} else {
		if data.Content == nil {
			return nil, nil
		}
		var ok bool
		content, ok = data.Content[language]
		if !ok {
			content = data.Content[10] // 取不到的情况下默认返回英语
		}
	}
	countdown := data.BeginTime - timer.UnixMilli()
	if countdown < 0 {
		countdown = 0
	}
	return &models.Announcement{
		Id:        data.Id,
		BeginTime: data.BeginTime,
		EndTime:   data.EndTime,
		StoreUrl:  data.StoreUrl,
		Content:   content,
		ShowLogin: data.ShowLogin,
		Countdown: countdown,
	}, nil
}

func (h *Announcement) Del(id string) error {
	a := model.NewAnnouncement()
	if err := a.DelPB(id); err != nil {
		return utils.NewDefaultErrorWithMsg("删除失败：" + err.Error())
	}
	if err := a.Del(id); err != nil {
		return utils.NewDefaultErrorWithMsg("删除失败：" + err.Error())
	}
	return nil
}

func (h *Announcement) getId(typ, version string) string {
	return typ + "@" + version
}
