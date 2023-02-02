package model

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

type Announcement struct {
	Id        string `db:"id" json:"id"`
	Type      string `db:"type" json:"type"`
	Version   string `db:"version" json:"version"`
	BeginTime int64  `db:"begin_time" json:"begin_time"`
	EndTime   int64  `db:"end_time" json:"end_time"`
	StoreUrl  string `db:"store_url" json:"store_url"`
	Content   string `db:"content" json:"content"`
	ShowLogin bool   `db:"show_login" json:"show_login"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

func NewAnnouncement() *Announcement {
	return &Announcement{}
}

func (a *Announcement) Save(data *Announcement) error {
	query := `INSERT INTO announcement (id,type,version,begin_time,end_time,store_url,content,show_login,created_at)
			VALUES (:id,:type,:version,:begin_time,:end_time,:store_url,:content,:show_login,:created_at)
			ON DUPLICATE KEY UPDATE type = VALUES(type),version = VALUES(version),begin_time = VALUES(begin_time),
			end_time = VALUES(end_time),store_url = VALUES(store_url),content = VALUES(content),show_login = VALUES(show_login);`

	if _, err := orm.GetMySQL().NamedExec(query, data); err != nil {
		return err
	}
	return nil
}

func (a *Announcement) Del(id string) error {
	query := `DELETE FROM announcement WHERE id=?`
	if _, err := orm.GetMySQL().Exec(query, id); err != nil {
		return err
	}
	return nil
}

func (a *Announcement) Find() ([]*Announcement, error) {
	query := "SELECT id,type,version,begin_time,end_time,store_url,content,show_login,created_at FROM announcement"
	dest := make([]*Announcement, 0)
	if err := orm.GetMySQL().Select(&dest, query); err != nil {
		return nil, err
	}
	return dest, nil
}

func (a *Announcement) GetPB(id string) (*dao.Announcement, *errmsg.ErrMsg) {
	data := &dao.Announcement{Id: id}
	ok, err := orm.GetOrm(ctx.GetContext()).GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, err
	}
	return data, nil
}

func (a *Announcement) SavePB(data *dao.Announcement) *errmsg.ErrMsg {
	db := orm.GetOrm(ctx.GetContext())
	db.SetPB(redisclient.GetDefaultRedis(), data)
	return db.Do()
}

func (a *Announcement) DelPB(id string) *errmsg.ErrMsg {
	data := &dao.Announcement{Id: id}
	db := orm.GetOrm(ctx.GetContext())
	db.DelPB(redisclient.GetDefaultRedis(), data)
	return db.Do()
}
