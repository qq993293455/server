package model

import (
	"context"
	"database/sql"
	"errors"

	"coin-server/common/ctx"
	commonmail "coin-server/common/mail"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	utils2 "coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/pikaviewer/utils"
)

type Player struct {
	RoleId     string `db:"role_id" json:"role_id,omitempty"`
	Nickname   string `db:"nickname" json:"nickname,omitempty"`
	CreateTime int64  `db:"create_time" json:"create_time,omitempty"`
	LoginTime  int64  `db:"login_time" json:"login_time,omitempty"`
}

func NewPlayer() *Player {
	return &Player{}
}

func (p *Player) CountByName(name string) (int, error) {
	query := "SELECT COUNT(*) num FROM game.roles WHERE nickname LIKE ?;"
	var count int
	err := orm.GetMySQL().Get(&count, query, name+"%")
	return count, err
}

func (p *Player) FindByName(name string) ([]*Player, error) {
	query := "SELECT role_id,nickname,create_time,login_time FROM game.roles WHERE nickname LIKE ? LIMIT 5;"
	data := make([]*Player, 0)
	err := orm.GetMySQL().Select(&data, query, name+"%")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *Player) GetRoles(roleIds []string) ([]*dao.Role, error) {
	ri := make([]orm.RedisInterface, len(roleIds))
	for idx := range ri {
		ri[idx] = &dao.Role{RoleId: roleIds[idx]}
	}
	notFound, err := orm.GetOrm(ctx.GetContext()).MGetPB(redisclient.GetUserRedis(), ri...)
	if err != nil {
		return nil, err
	}
	list := make([]*dao.Role, 0)
	notFoundMap := make(map[int]struct{}, len(notFound))
	for _, i := range notFound {
		notFoundMap[i] = struct{}{}
	}
	for i := range ri {
		if _, ok := notFoundMap[i]; ok {
			continue
		}
		list = append(list, ri[i].(*dao.Role))
	}
	return list, nil
}

func (p *Player) GetServerId(userIds []string) (map[string]values.ServerId, error) {
	ri := make([]orm.RedisInterface, len(userIds))
	for idx := range ri {
		ri[idx] = &dao.User{UserId: userIds[idx]}
	}
	notFound, err := orm.GetOrm(ctx.GetContext()).MGetPB(redisclient.GetUserRedis(), ri...)
	if err != nil {
		return nil, err
	}
	notFoundMap := make(map[int]struct{}, len(notFound))
	for _, i := range notFound {
		notFoundMap[i] = struct{}{}
	}
	ret := make(map[string]values.ServerId)

	for i := range ri {
		if _, ok := notFoundMap[i]; ok {
			continue
		}
		user := ri[i].(*dao.User)
		ret[user.RoleId] = user.ServerId
	}
	return ret, nil
}

func (p *Player) GetRole(roleId values.RoleId) (*dao.Role, error) {
	role := &dao.Role{RoleId: roleId}
	ok, err := orm.GetOrm(ctx.GetContext()).GetPB(redisclient.GetUserRedis(), role)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return role, nil
}

func (p *Player) GetOneServerId(userId string) (values.ServerId, error) {
	user := &dao.User{UserId: userId}
	ok, err := orm.GetOrm(ctx.GetContext()).GetPB(redisclient.GetUserRedis(), user)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}
	return user.ServerId, nil
}

func (p *Player) GetUserData(userId string) (*dao.User, bool, error) {
	user := &dao.User{UserId: userId}
	ok, err := orm.GetOrm(ctx.GetContext()).GetPB(redisclient.GetUserRedis(), user)
	if err != nil {
		return nil, false, utils.NewDefaultErrorWithMsg(err.Error())
	}
	if !ok {
		return nil, false, nil
	}
	return user, true, nil
}

func (p *Player) EntireMail() ([]*dao.MailItem, error) {
	mail := &dao.EntireMail{Key: commonmail.EntireMailKey}
	ok, err := orm.GetOrm(ctx.GetContext()).GetPB(commonmail.GetMailRedis(), mail)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return mail.Mails, nil
}

func (p *Player) SaveEntireMail(mails []*dao.MailItem) error {
	mail := &dao.EntireMail{
		Key:   commonmail.EntireMailKey,
		Mails: mails,
	}
	db := orm.GetOrm(ctx.GetContext())
	db.SetPB(commonmail.GetMailRedis(), mail)
	if err := db.Do(); err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}

func (p *Player) Ban(user *dao.User) error {
	db := orm.GetOrm(ctx.GetContext())
	db.SetPB(redisclient.GetUserRedis(), user)
	if err := db.Do(); err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}

func (p *Player) Total() (int64, error) {
	count, err := redisclient.GetDefaultRedis().Get(context.Background(), redisclient.RegisterCount).Int64()
	if err == redisclient.Nil {
		err = nil
	}
	return count, err
}

func (p *Player) GetItemCount(roleId values.RoleId, id ...values.ItemId) (map[values.ItemId]values.Integer, error) {
	if len(id) <= 0 {
		return nil, nil
	}
	items := make([]orm.RedisInterface, 0)
	for _, itemId := range id {
		items = append(items, &dao.Item{ItemId: itemId})
	}
	db := orm.GetOrm(ctx.GetContext())
	_, err := db.HMGetPB(redisclient.GetUserRedis(), utils2.GenDefaultRedisKey(values.ItemBag, values.Hash, roleId), items)
	if err != nil {
		return nil, utils.NewDefaultErrorWithMsg(err.ErrMsg)
	}
	ret := make(map[values.ItemId]values.Integer)
	for _, el := range items {
		item, ok := el.(*dao.Item)
		if !ok {
			continue
		}
		ret[item.ItemId] += item.Count
	}
	return ret, nil
}
