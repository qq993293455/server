package ipregion

import (
	_ "embed"
	"strings"

	"coin-server/common/utils"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

//go:embed ip2region.xdb
var fs []byte

var searcher *xdb.Searcher

func Init() {
	var err error
	searcher, err = xdb.NewWithBuffer(fs)
	utils.Must(err)
}

// IsNorthAmerica 是否是北美IP
func IsNorthAmerica(ip string) bool {
	ret, err := searcher.SearchByStr(ip)
	if err != nil {
		return false
	}
	return strings.HasPrefix(ret, "美国") ||
		strings.HasPrefix(ret, "加拿大") ||
		strings.HasPrefix(ret, "美属萨摩亚") ||
		strings.HasPrefix(ret, "关岛") ||
		strings.HasPrefix(ret, "马绍尔群岛") ||
		strings.HasPrefix(ret, "北马里亚纳群岛") ||
		strings.HasPrefix(ret, "帕劳") ||
		strings.HasPrefix(ret, "美属维尔京群岛")
}
