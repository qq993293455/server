package bufferstream

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type MMM struct {
	A int64
	B string
}

func TestWrite(t *testing.T) {
	r := require.New(t)
	_ = r
	var bs BaseStream
	i8 := int8(7)
	Write(&bs, i8)
	str := "123123"
	Write(&bs, str)
	a := [4]int64{1, 2, 3, 4}
	Write(&bs, a)
	slice := []byte("44411asca萨达")
	Write(&bs, slice)

	slice1 := []string{"44411asca萨达", "232ss"}
	Write(&bs, slice1)

	slice2 := [][]string{{"mmmmmmmm", "nnnnnn"}, {"lllllll"}}
	Write(&bs, slice2)
	m := map[int]int{1: 1, 2: 2, 4444444444: 2822727627}
	Write(&bs, &m)
	bool1 := true
	Write(&bs, bool1)
	mmm := &MMM{}
	Write(&bs, mmm)
	data, err := bs.ToLZ4()
	fmt.Println(bs.len_, len(data), err)
}
