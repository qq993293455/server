package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var key = []byte("9yvX@0IXQ*6_VICCGWpGyS0DpXBWIekv")

type claims struct {
	*jwt.StandardClaims
	*ClaimsInfo
}

type ClaimsInfo struct {
	UId  int64
	Role int64
}

func GenToken(uid, role int64, duration time.Duration) (string, error) {
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(duration).Unix(),
		},
		ClaimsInfo: &ClaimsInfo{
			UId:  uid,
			Role: role,
		},
	})
	token, err := tokenObj.SignedString(key)
	if err != nil {
		return "", err
	}
	return token, nil
}

func ParseToken(token string) (*ClaimsInfo, error) {
	tokenObj, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (i interface{}, err error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := tokenObj.Claims.(*claims)
	if !ok || !tokenObj.Valid {
		return nil, NewDefaultErrorWithMsg("登录已过期")
	}
	return c.ClaimsInfo, nil
}

func SetToken(ctx *gin.Context, token string) {
	ctx.Set("Token", token)
}

func GetToken(ctx *gin.Context) (string, bool) {
	t, ok := ctx.Get("Token")
	if !ok {
		return "", false
	}
	token, ok := t.(string)
	if !ok {
		return "", false
	}
	return token, true
}
