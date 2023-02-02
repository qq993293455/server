package trans

import (
	"unsafe"
)

func SliceDao2Model[D any, M any](dao D) M {
	return *(*M)(unsafe.Pointer(&dao))
}

func SliceModel2Dao[M any, D any](model M) D {
	return *(*D)(unsafe.Pointer(&model))
}
