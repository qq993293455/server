package utils

import (
	"crypto/md5"
	"fmt"
)

func MD5(data []byte) string {
	srcCode := md5.Sum(data)
	return fmt.Sprintf("%x", srcCode)
}

func MD5String(str string) string {
	return MD5(StringToBytes(str))
}
