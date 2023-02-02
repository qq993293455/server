package stage

import (
	"unsafe"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
)

func ExploreModel2Dao(exploreModel *models.StageExplore) *dao.StageExplore {
	return (*dao.StageExplore)(exploreModel)
}

func ExploreDao2Model(exploreDao *dao.StageExplore) *models.StageExplore {
	return (*models.StageExplore)(exploreDao)
}

func ExploreModels2Dao(exploreModel []*models.StageExplore) []*dao.StageExplore {
	return *(*[]*dao.StageExplore)(unsafe.Pointer(&exploreModel))
}

func ExploreDao2Models(exploreDao []*dao.StageExplore) []*models.StageExplore {
	return *(*[]*models.StageExplore)(unsafe.Pointer(&exploreDao))
}
