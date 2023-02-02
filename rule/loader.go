package rule

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"coin-server/common/logger"
	"coin-server/common/utils"
	rule_model "coin-server/rule/rule-model"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Request struct {
	Code   int    `json:"code"`
	Error  string `json:"error"`
	Result struct {
		Branch  string      `json:"branch"`
		Version int64       `json:"version"`
		Rule    []TableData `json:"rule"`
	} `json:"result"`
}

type TableData struct {
	Table string `json:"table"`
	Data  string `json:"data"`
}

func LoadRule() {
	rule := utils.GetRuleName()
	logger.DefaultLogger.Info("loading latest rule", zap.String("branch", rule))
	req := NewRequest()
	if err := req.Get(rule, ""); err != nil {
		panic(err)
	}
	v := viper.New()
	version := strconv.Itoa(int(req.Result.Version))
	v.Set(rule_model.VersionKey, version)
	for _, data := range req.Result.Rule {
		v.Set(data.Table, data.Data)
	}
	rule_model.Parsing(rule_model.NewData(v), rule_model.NewCustomParse())
}

func LoadRuleByFile() {
	dir := "../../../rule/rules/"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	v := viper.New()
	for _, f := range files {
		b, err := ioutil.ReadFile(dir + f.Name())
		if err != nil {
			panic(err)
		}
		if f.Name() != "version.json" {
			v.Set(strings.Split(f.Name(), ".")[0], string(b))
		}
	}
	rule_model.Parsing(rule_model.NewData(v), rule_model.NewCustomParse())
}

func NewRequest() *Request {
	return &Request{}
}

func (r *Request) Get(name, version string) error {
	resp, err := http.Get(fmt.Sprintf("%s?branch=%s&version=%s", addr, name, version))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, r); err != nil {
		return err
	}
	if r.Code != 200 {
		panic(fmt.Errorf("获取规则失败，code = %d, err = %s", r.Code, r.Error))
	}
	return nil
}
