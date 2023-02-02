package ipregion

import (
	"fmt"
	"testing"
)

func init() {
	Init()
}

func TestIsAmerica(t *testing.T) {
	fmt.Println(searcher.SearchByStr("23.128.224.255"))
	fmt.Println(searcher.SearchByStr("66.249.69.164"))
	fmt.Println(searcher.SearchByStr("204.101.161.159"))
	fmt.Println(IsNorthAmerica("66.249.69.164"))
	fmt.Println(IsNorthAmerica("204.101.161.159"))
}

func BenchmarkIsAmerica(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsNorthAmerica("66.249.69.164")
	}
}
