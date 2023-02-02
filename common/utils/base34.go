package utils

import "math"

var baseStr = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"
var baseLen = uint64(len(baseStr))
var baseMap [math.MaxUint8]uint64

func init() {
	for i, v := range StringToBytes(baseStr) {
		baseMap[v] = uint64(i)
	}
}
func Base34EncodeToString(id uint64) string {
	return BytesToString(Base34Encode(id))
}

func Base34Encode(id uint64) []byte {
	quotient := id
	mod := uint64(0)
	l := make([]byte, 0, 13)
	for quotient != 0 {
		mod = quotient % baseLen
		quotient = quotient / baseLen
		l = append(l, baseStr[int(mod)])
	}
	ls := len(l)
	for i := 0; i < ls/2; i++ {
		l[i], l[ls-1-i] = l[ls-1-i], l[i]
	}
	if ls == 0 {
		l = append(l, '0')
	}
	return l
}

func Base34Decode(bs []byte) uint64 {
	ls := len(bs)
	if ls == 0 {
		return 0
	}
	var res, r uint64
	for i := ls - 1; i >= 0; i-- {
		v := baseMap[bs[i]]
		var b uint64 = 1
		for j := uint64(0); j < r; j++ {
			b *= baseLen
		}
		res += b * v
		r++
	}
	return res
}

func Base34DecodeString(str string) uint64 {
	return Base34Decode(StringToBytes(str))
}
