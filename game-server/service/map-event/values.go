package map_event

import (
	"unsafe"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
)

const (
	TreasureChestID = int64(1)
	CollectMineID   = int64(2)
	BlackMarketID   = int64(3)
	JigsawPuzzleID  = int64(4)
	CardMatchID     = int64(5)
	WantingID       = int64(6)
	MeetID          = int64(7)
	BuildArmsID     = int64(8)

	AnecdotesNum     = "AnecdotesNum"
	AnecdotesAddTime = "AnecdotesAddTime"
)

func DecodeBit0(bitmap int64, len int64) []int64 {
	p := make([]int64, 0)
	for i := int64(1); i < len+1; i++ {
		if bitmap&(1<<(i-1)) == 0 {
			p = append(p, i)
		}
	}
	return p
}

func DecodeBit1(bitmap int64, len int64) []int64 {
	p := make([]int64, 0)
	for i := int64(1); i <= len; i++ {
		if bitmap&(1<<(i-1)) != 0 {
			p = append(p, i)
		}
	}
	return p
}

func EncodeBit1(bitmap int64, id int64) int64 {
	return bitmap + 1<<(id-1)
}

func StoryModels2Dao(storyModel []*models.MapStory) []*dao.MapStory {
	return *(*[]*dao.MapStory)(unsafe.Pointer(&storyModel))
}

func StoryDao2Models(storyDao []*dao.MapStory) []*models.MapStory {
	return *(*[]*models.MapStory)(unsafe.Pointer(&storyDao))
}
