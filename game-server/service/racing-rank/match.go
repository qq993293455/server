package racing_rank

import (
	"errors"
	"time"

	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/game-server/service/racing-rank/dao"
)

const maxCallCount = 5

func (svc *Service) matching(self *pbdao.RacingRankMatch) error {
	count := int(self.Count)
	var err error
	ltList := make([]*dao.Data, 0)
	ltList, err = svc.fetchFormMySQL(self.CombatValue, true, ltList, count, 0)
	if err != nil {
		return err
	}
	gtList := make([]*dao.Data, 0)
	gtList, err = svc.fetchFormMySQL(self.CombatValue, false, gtList, count, 0)
	if err != nil {
		return err
	}
	ltCount := 9
	gtCount := 40
	minTime := time.Now().AddDate(0, 0, -7).Unix()
	list := make([]*models.RankItem, 0)
	tempMap := make(map[values.RoleId]struct{})
	// 优先取7日内登录过的玩家
	for _, item := range ltList {
		if item.LoginTime >= minTime {
			list = append(list, &models.RankItem{
				RoleId: item.RoleId,
			})
			tempMap[item.RoleId] = struct{}{}
			ltCount--
			if ltCount <= 0 {
				break
			}
		}
	}
	for _, item := range gtList {
		if item.LoginTime >= minTime {
			list = append(list, &models.RankItem{
				RoleId: item.RoleId,
			})
			tempMap[item.RoleId] = struct{}{}
			gtCount--
			if gtCount <= 0 {
				break
			}
		}
	}
	// 填补低于当前战力的数据
	if ltCount > 0 {
		for _, item := range ltList {
			if _, ok := tempMap[item.RoleId]; !ok {
				list = append(list, &models.RankItem{
					RoleId: item.RoleId,
				})
				tempMap[item.RoleId] = struct{}{}
				ltCount--
				if ltCount <= 0 {
					break
				}
			}
		}
	}
	// 填补高于当前战力的数据
	if gtCount > 0 {
		for _, item := range gtList {
			if _, ok := tempMap[item.RoleId]; !ok {
				list = append(list, &models.RankItem{
					RoleId: item.RoleId,
				})
				tempMap[item.RoleId] = struct{}{}
				gtCount--
				if gtCount <= 0 {
					break
				}
			}
		}
	}
	// 可能会存在不够填补的情况，直接不区分低于或高于的，直接填补够需要的数量
	if len(list) < count {
		tempList := make([]*dao.Data, 0)
		tempList = append(tempList, ltList...)
		tempList = append(tempList, gtList...)
		for _, item := range tempList {
			if _, ok := tempMap[item.RoleId]; !ok {
				list = append(list, &models.RankItem{
					RoleId: item.RoleId,
				})
				tempMap[item.RoleId] = struct{}{}
				if len(list) >= count {
					break
				}
			}
		}
	}
	list = append(list, self.Self)
	if err := dao.SaveDataImmediately(&pbdao.RacingRankData{
		RoleId:       self.RoleId,
		List:         list,
		Locked:       false,
		ForceRefresh: true, // 将强制更新置为true
	}); err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func (svc *Service) fetchFormMySQL(combatValue values.Integer, lt bool, list []*dao.Data, count, callCount int) ([]*dao.Data, error) {
	data, err := dao.Find(combatValue, lt)
	if err != nil {
		return list, err
	}
	if len(data) <= 0 {
		return list, nil
	}
	list = append(list, data...)
	callCount++
	if callCount >= maxCallCount || len(list) >= count {
		return list, nil
	}
	combatValue = data[len(data)-1].HighestPower

	time.Sleep(time.Millisecond * 100)
	return svc.fetchFormMySQL(combatValue, lt, list, count, callCount)
}
