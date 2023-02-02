package jwt

import (
	"time"
	"unsafe"

	"coin-server/common/timer"

	"github.com/rs/xid"

	"github.com/golang-jwt/jwt"
)

type ClaimData struct {
	RoleId      string `json:"role_id"`
	RuleVersion string `json:"rule_version"`
	ServerId    int64  `json:"server_id"`
}

type RoleIdClaims struct {
	Data *ClaimData `json:"data"`
	jwt.StandardClaims
}

var jwtKey = []byte("iggcdl5,.")

func SignJwt(roleId string, ruleVersion string, serverId int64) (string, error) {
	now := timer.Now()
	c := &RoleIdClaims{
		Data: &ClaimData{
			RoleId:      roleId,
			RuleVersion: ruleVersion,
			ServerId:    serverId,
		},
		StandardClaims: jwt.StandardClaims{
			Audience:  "app",
			ExpiresAt: now.Add(time.Hour * 24).Unix(),
			Id:        xid.NewWithTime(now).String(),
			IssuedAt:  now.Unix(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString(jwtKey)
}

func ParseJwt(b []byte) (*ClaimData, error) {
	c := &RoleIdClaims{}
	_, err := jwt.ParseWithClaims(*(*string)(unsafe.Pointer(&b)), c, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	return c.Data, nil
}
