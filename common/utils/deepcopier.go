package utils

import (
	"github.com/antlabs/deepcopy"
)

// 通过反射拷贝
func deepCopy(dst interface{}, src interface{}) {
	Must(deepcopy.Copy(dst, src).Do())
}

// 深拷贝
func DeepCopy(dst interface{}, src interface{}) {
	deepCopy(dst, src)
}
