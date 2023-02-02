package controller

import (
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

type Routes struct {
}

type Item struct {
	Router   string   `json:"router"`
	Children []string `json:"children"`
}

var (
	operator = []string{"mail", "announcement", "battleLog", "version", "broadcast", "whiteList", "cdkey", "beta", "mapLineInfo", "player"}
	guest    = []string{"pika", "restart", "build", "rule", "consul", "NATSBoard"}
	gitlab   = "gitlab"
)

func (c *Routes) List(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	token, ok := utils.GetToken(ctx)
	if !ok {
		resp.Send(utils.NewDefaultErrorWithMsg("登录已过期，请重新登录"))
		return
	}
	info, err := utils.ParseToken(token)
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send([]interface{}{gin.H{
		"router":   "root",
		"children": getRouteByRole(info.Role),
	}})
}

func getRouteByRole(role int64) []*Item {
	list := make([]*Item, 0)
	all := make([]string, 0)
	switch role {
	case utils.SuperAdmin:
		all = append(operator, guest...)
		all = append(all, gitlab)
	case utils.Admin:
		all = append(operator, guest...)
	case utils.Guest:
		all = guest
	case utils.Operator:
		all = operator
	}
	for _, el := range all {
		item := &Item{
			Router:   el,
			Children: nil,
		}
		if el == "mail" {
			item.Children = []string{"query", "send"}
		}
		list = append(list, item)
	}
	return list
}
