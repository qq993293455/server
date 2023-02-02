package maintask

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

func SetMainTaskFinish(finish *models.MainTaskFinish, chapter, idx values.Integer, status models.RewardStatus) {
	if finish == nil {
		return
	}
	if finish.ChapterFinish == nil {
		finish.ChapterFinish = map[int64]*models.MainTaskChapterFinish{}
	}

	if v, ok := finish.ChapterFinish[chapter]; ok {
		if status == models.RewardStatus_Unlocked &&
			v.Finish[idx] != models.RewardStatus_Locked {
			return
		}
		v.Finish[idx] = status
	} else {
		finish.ChapterFinish[chapter] = &models.MainTaskChapterFinish{Finish: map[int64]models.RewardStatus{idx: status}}
	}
}
