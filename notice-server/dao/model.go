package dao

type Notice struct {
	Id            string `db:"id" json:"id"`
	Title         string `db:"title" json:"title"`
	Content       string `db:"content" json:"content"`
	RewardContent string `db:"reward_content" json:"reward_content"`
	IsCustom      bool   `db:"is_custom" json:"is_custom"`
	BeginAt       int64  `db:"begin_at" json:"begin_at"`
	ExpiredAt     int64  `db:"expired_at" json:"expired_at"`
	CreatedAt     int64  `db:"created_at" json:"created_at"`
}
