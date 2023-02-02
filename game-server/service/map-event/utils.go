package map_event

import (
	"coin-server/common/values"

	"github.com/golang-jwt/jwt"
)

var JwtKey = []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")

type Claims struct {
	TokenInfo
	jwt.StandardClaims
}

type TokenInfo struct {
	RoleId values.RoleId
	Option values.Integer
}
