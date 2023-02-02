package handler

import (
	"coin-server/pikaviewer/env"
	"io/ioutil"
	"os"
	"strings"

	"coin-server/pikaviewer/utils"
)

type BattleLog struct {
}

func (h *BattleLog) FileList() ([]*utils.Files, error) {
	fsInfo, err := ioutil.ReadDir(os.Getenv(env.BATTLE_LOG_DIR))
	if err != nil {
		return nil, err
	}
	files := make([]*utils.Files, 0)
	for _, info := range fsInfo {
		if !info.IsDir() && strings.Index(info.Name(), "fight") != -1 {
			if info.Size() <= 0 {
				continue
			}
			files = append(files, &utils.Files{
				Name: info.Name(),
				Size: info.Size(),
				Time: info.ModTime(),
			})
		}
	}
	return files, nil
}

func (h *BattleLog) Delete(list []string) error {
	for _, name := range list {
		if err:= os.Remove(os.Getenv(env.BATTLE_LOG_DIR) + "/" + name);err!=nil{
			return err
		}
	}
	return nil
}
