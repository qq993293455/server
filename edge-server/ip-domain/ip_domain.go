package ipdomain

import "coin-server/common/consulkv"

var ipMap = map[string]string{}

func Init(cnf *consulkv.Config) {
	err := cnf.Unmarshal("ip-domain", &ipMap)
	if err != nil {
		panic(err)
	}
}
func IPToDomain(str string) string {
	v, ok := ipMap[str]
	if ok {
		return v
	}
	return str
}
