package model

import (
	"database/sql"
	"errors"

	"coin-server/common/orm"
)

type User struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Role      int64  `json:"role"`
	Status    int64  `json:"status"`
	CreatedAt int64  `json:"createdAt"`
}

func NewUser() *User {
	return &User{}
}

func (u *User) GetByName(name string) (bool, error) {
	query := "SELECT id,username,password,role,status FROM admin_user WHERE username=?"
	err := orm.GetMySQL().Get(u, query, name)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
