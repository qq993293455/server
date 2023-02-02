package utils

import (
	"testing"
)

const count = uint64(100000000)

func TestBase34(t *testing.T) {
	//r := require.New(t)
	//id := uint64(math.MaxUint64)
	//c := Base34Encode(id)
	//fmt.Println(string(c))
	//id1 := Base34Decode([]byte("13RR4"))
	//fmt.Println(id1)
	//r.Equal(id, id1)
	//
	t.Log(Base34EncodeToString(1283021))
	m := make(map[string]uint64, count)

	for i := uint64(0); i < count; i++ {
		c := Base34EncodeToString(i)
		if _, ok := m[c]; ok {
			panic(i)
		} else {
			m[c] = i
		}
	}
}

func BenchmarkBase34Decode(b *testing.B) {
	id := uint64(1)
	c := Base34EncodeToString(id)
	for i := 0; i < b.N; i++ {
		Base34DecodeString(c)
	}
}
