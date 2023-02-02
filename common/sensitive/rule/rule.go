package rule

import (
	"strings"

	"coin-server/rule"
)

func Dict1Words() []string {
	dict := rule.MustGetReader(nil).SensitiveDict1.List()
	ret := make([]string, len(dict))
	for i, v := range dict {
		ret[i] = strings.ToLower(v.Word)
	}
	return ret
}

func Dict2Words() []string {
	dict := rule.MustGetReader(nil).SensitiveDict2.List()
	ret := make([]string, len(dict))
	for i, v := range dict {
		ret[i] = strings.ToLower(v.Word)
	}
	return ret
}
