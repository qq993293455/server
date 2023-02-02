package v1

import (
	"encoding/json"
	"math"
	"reflect"
	"sort"
	"strconv"
	time2 "time"

	"coin-server/common/logger"
	utils2 "coin-server/common/utils"
	"coin-server/pikaviewer/controller"
	"coin-server/pikaviewer/global"
	"coin-server/pikaviewer/utils"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"
)

var (
	payCtrl        = new(controller.Pay)
	gameNoticeCtrl = new(controller.GameNotice)
)

func ApiRouter(r *gin.Engine) {
	api := r.Group("api")
	api.Use(verifySign(), verifyIP())
	v1 := api.Group("v1")

	pay := v1.Group("pay")
	pay.POST("callback", payCtrl.PaySuccess)

	mail := v1.Group("mail")
	mail.POST("send", mailCtrl.Send)

	player := v1.Group("player")
	player.POST("info", playerCtrl.GetPlayerInfoForSOP)
	player.POST("ban", playerCtrl.BanPlayer)
	player.POST("kick", playerCtrl.KickPlayer)
	player.POST("chat/ban", playerCtrl.BanChat)
	player.POST("chat/unban", playerCtrl.UnBanChat)
	player.POST("online", playerCtrl.OnlineCount)
	player.POST("total", playerCtrl.Total)
	player.POST("update/currency", playerCtrl.UpdatePlayerCurrency)

	notice := v1.Group("notice")
	notice.POST("send", noticeCtrl.Broadcast)
	notice.POST("publish", gameNoticeCtrl.Publish)
	notice.POST("delete", gameNoticeCtrl.Delete)
}

func verifySign() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !global.Config.EnableSign {
			ctx.Next()
			return
		}
		response := utils.NewResponse(ctx)
		obj := make(map[string]interface{})
		if err := ctx.ShouldBindBodyWith(&obj, binding.JSON); err != nil {
			ctx.Abort()
			response.Send(utils.InvalidArguments)
			return
		}
		logger.DefaultLogger.Debug("SOP API Request data",
			zap.String("uri", ctx.Request.RequestURI),
			zap.Any("data", obj),
		)
		params := make(map[string]string)
		keys := make([]string, 0)
		var (
			sign  string
			time  int64
			debug bool
		)
		for k, v := range obj {
			tk := reflect.TypeOf(v).Kind()
			switch tk {
			case reflect.String:
				val := v.(string)
				params[k] = val
				if k == "sign" {
					sign = val
				} else {
					keys = append(keys, k)
				}
			case reflect.Float64:
				val := v.(float64)
				params[k] = strconv.Itoa(int(val))
				if k == "time" {
					time = int64(val) // fmt.Sprintf("%1.0f", v)
				}
				keys = append(keys, k)
			case reflect.Bool:
				val := v.(bool)
				if k == "debug" {
					debug = val
				} else {
					boolVal := "false"
					if val {
						boolVal = "true"
					}
					params[k] = boolVal
					keys = append(keys, k)
				}
			default:
				b, err := json.Marshal(v)
				if err != nil {
					ctx.Abort()
					response.Send(utils.NewErrWithCodeMsg(utils.InvalidArguments, k+": "+err.Error()))
					return
				}
				params[k] = string(b)
				keys = append(keys, k)
			}
		}
		if math.Abs(float64(time-time2.Now().Unix())) > 600000 {
			ctx.Abort()
			response.Send(utils.NewErrWithCodeMsg(utils.InvalidArguments, "invalid arguments: time"))
			return
		}
		if sign == "" {
			ctx.Abort()
			response.Send(utils.NewErrWithCodeMsg(utils.InvalidArguments, "invalid arguments: sign"))
			return
		}
		sort.Strings(keys)
		var plaintext string
		for _, key := range keys {
			val := params[key]
			if val == "" {
				ctx.Abort()
				response.Send(utils.NewErrWithCodeMsg(utils.InvalidArguments, "invalid arguments: "+key))
				return
			}
			plaintext += key + "=" + val + "&"
		}
		if len(plaintext) > 0 {
			plaintext = plaintext[0 : len(plaintext)-1]
		}
		_sign := utils2.MD5String(plaintext + global.Config.SignKey)

		if debug {
			ctx.Set(utils.DebugKey, true)
			ctx.Set(utils.DebugSignPlaintextKey, plaintext)
			ctx.Set(utils.DebugSignKey, _sign)
		}

		if sign != _sign {
			ctx.Abort()
			response.Send(utils.InvalidSign)
			return
		}
		ctx.Next()
	}
}

func verifyIP() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ipList := global.Config.IPWhiteList
		if len(ipList) == 0 {
			ctx.Next()
			return
		}
		clientIP := ctx.ClientIP()
		var exist bool
		for _, ip := range ipList {
			if clientIP == ip {
				exist = true
				break
			}
		}
		if !exist {
			response := utils.NewResponse(ctx)
			ctx.Abort()
			response.Send(utils.IllegalRequest)
			return
		}
		ctx.Next()
	}
}
