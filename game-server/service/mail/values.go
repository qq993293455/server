package mail

import (
	"encoding/json"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/common/values/enum"
)

// 锁
const (
	mailLock             = "lock:mail:"
	entireServerMailLock = "lock:entire:server"
	updateRewardMailLock = "lock:update:reward"
)

func model2daoMailItem(mail *models.Mail, read bool) *dao.MailItem {
	return &dao.MailItem{
		Id:          mail.Id,
		Type:        mail.Type,
		Sender:      mail.Sender,
		TextId:      mail.TextId,
		Title:       mail.Title,
		Content:     mail.Content,
		Hi:          mail.Hi,
		ExpiredAt:   mail.ExpiredAt,
		Args:        mail.Args,
		Attachment:  mail.Attachment,
		Read:        read,
		CreatedAt:   mail.CreatedAt,
		ActivatedAt: mail.ActivatedAt,
	}
}

func daoItem2model(mail *dao.MailItem) *models.Mail {
	return &models.Mail{
		Id:          mail.Id,
		Type:        mail.Type,
		Sender:      mail.Sender,
		TextId:      mail.TextId,
		Title:       mail.Title,
		Content:     mail.Content,
		Hi:          mail.Hi,
		ExpiredAt:   mail.ExpiredAt,
		Args:        mail.Args,
		Attachment:  mail.Attachment,
		CreatedAt:   mail.CreatedAt,
		ActivatedAt: mail.ActivatedAt,
	}
}

func entire2user(item *dao.MailItem, language values.Integer) *dao.MailItem {
	title := item.Title
	titleMap := make(map[int64]string)
	if err := json.Unmarshal([]byte(item.Title), &titleMap); err == nil {
		var ok bool
		title, ok = titleMap[language]
		if !ok {
			title = titleMap[enum.DefaultLanguage]
		}
	}

	content := item.Content
	contentMap := make(map[int64]string)
	if err := json.Unmarshal([]byte(item.Content), &contentMap); err == nil {
		var ok bool
		content, ok = contentMap[language]
		if !ok {
			content = contentMap[enum.DefaultLanguage]
		}
	}

	return &dao.MailItem{
		Id:          "entire@" + item.Id,
		Type:        item.Type,
		Sender:      item.Sender,
		TextId:      item.TextId,
		Title:       title,
		Content:     content,
		Hi:          item.Hi,
		ExpiredAt:   item.ExpiredAt,
		Args:        item.Args,
		Attachment:  item.Attachment,
		Read:        false,
		CreatedAt:   item.CreatedAt,
		DeletedAt:   item.DeletedAt,
		ActivatedAt: item.ActivatedAt,
	}
}
