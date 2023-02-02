package iggreq

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// QueryGeneratorInterface Generate query string.
type QueryGeneratorInterface interface {
	// GenQueryString generate query string.
	GenQueryString() string
	// ContentType content-type of query data for POST request
	ContentType() string
}

// queryParam Paramiter structure for internal use
type queryParam struct {
	Name  string
	Value string
}

// QueryMaker stores parameters and makes HTTP query body
type QueryMaker struct {
	// key for sign querystring
	SignKey string
	// Parameters store
	Params []queryParam
}

// NewQueryMaker create a new query
func NewQueryMaker(scode string) *QueryMaker {
	return &QueryMaker{
		SignKey: scode,
	}
}

// AddParam adds one parameter to query
func (q *QueryMaker) AddParam(name string, value interface{}) *QueryMaker {
	q.Params = append(q.Params, queryParam{
		Name:  name,
		Value: fmt.Sprintf("%v", value),
	})
	return q
}

// GenQueryString generate query string.
// The final query string need an additional field named "sign".
// It's calculated as folowing:
// First generate query string from parameters without URL encoding.
// Next adds the string with security code.
// Then Do MD5 sum to the result string, and get HEX string form of sum
// Finally append the field "sign" with that string
func (q *QueryMaker) GenQueryString() string {
	// 排序
	sort.Slice(q.Params, func(i, j int) bool {
		return strings.Compare(q.Params[i].Name, q.Params[j].Name) < 0
	})

	// 构造
	sb := strings.Builder{}
	res := strings.Builder{}
	for _, p := range q.Params {
		sb.WriteString(fmt.Sprintf("%s=%s&", p.Name, p.Value))
		res.WriteString(fmt.Sprintf("%s=%s&", p.Name, url.QueryEscape(p.Value)))
	}

	// 签名
	if q.SignKey != "" {
		query := strings.TrimSuffix(sb.String(), "&")
		md5 := md5.Sum([]byte(query + q.SignKey))
		res.WriteString(fmt.Sprintf("sign=%v", hex.EncodeToString(md5[:])))
	}

	return strings.TrimSuffix(res.String(), "&")
}

// ContentType content-type of query data for POST request
func (q *QueryMaker) ContentType() string {
	return "application/x-www-form-urlencoded"
}
