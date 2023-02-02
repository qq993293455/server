package sensitive

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"coin-server/common/consulkv"
	"coin-server/common/sensitive/rule"

	"github.com/importcjj/sensitive"
)

var (
	fuzzy  = sensitive.New()       // 模糊匹配 表：dict2
	equals = map[string]struct{}{} // 精确匹配 表：dict1
)

func Init(conf *consulkv.Config) {
	fuzzy.AddWord(rule.Dict2Words()...)

	for _, word := range rule.Dict1Words() {
		equals[strings.ToLower(word)] = struct{}{}
	}
}

// 是否是符号
func isSymbol(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsNumber(r)
}

// ChatReplace 替换敏感词为*，不区分大小写
func ChatReplace(str string) string {
	lower := strings.ToLower(str)
	repl := fuzzy.Replace(lower, '*')

	runes := []rune(str)
	for i, word := range []rune(repl) {
		if word == '*' {
			runes[i] = '*'
		}
	}

	words := strings.FieldsFunc(repl, isSymbol) // 简单分词

	idx := 0
	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)
		if _, ok := equals[word]; ok {
			for j := idx; j < idx+wordLen; j++ {
				runes[j] = '*'
			}
		}
		idx += wordLen + 1
	}
	return string(runes)
}

func TextValid(str string) bool {
	lower := strings.ToLower(str)
	valid, _ := fuzzy.Validate(lower)
	if !valid {
		return valid
	}

	words := strings.FieldsFunc(lower, isSymbol) // 简单分词
	for _, word := range words {
		if _, ok := equals[word]; ok {
			return false
		}
	}

	return true
}
