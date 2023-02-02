//go:build linux
// +build linux

package service

import icUtils "coin-server/pikaviewer/utils"

func send2IC(tag string, params map[string]string) error {
	data := map[string]string{
		"token":        "df8a445dd467a62bf1d7bdc5066dd918",
		"target":       "group",
		"room":         "10073164",
		"title":        "战斗集群服务(EDGE):" + tag,
		"content_type": "1",
	}
	for k, v := range params {
		data[k] = v
	}
	_, err := icUtils.NewRequest("http://im-api.skyunion.net/msg").Post(data)
	return err
}
