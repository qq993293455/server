package atlas

import (
	"unsafe"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
)

func AtlasModels2Dao(atlasModel []*models.Atlas) []*dao.Atlas {
	return *(*[]*dao.Atlas)(unsafe.Pointer(&atlasModel))
}

func AtlasDao2Models(atlasDao []*dao.Atlas) []*models.Atlas {
	return *(*[]*models.Atlas)(unsafe.Pointer(&atlasDao))
}

func AtlasModel2Dao(atlasModel *models.Atlas) *dao.Atlas {
	return (*dao.Atlas)(atlasModel)
}

func AtlasDao2Model(atlasDao *dao.Atlas) *models.Atlas {
	return (*models.Atlas)(atlasDao)
}
