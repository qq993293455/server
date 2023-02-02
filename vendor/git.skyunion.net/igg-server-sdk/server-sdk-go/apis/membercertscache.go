package apis

import (
	"crypto/rsa"
	"encoding/json"
	"math/big"
	"time"

	"git.skyunion.net/igg-server-sdk/server-sdk-go/iggreq"
)

type certsCache struct {
	rsakeyMap  map[string]rsa.PublicKey
	lastUpdate int64
}

func (c *certsCache) Add(rkid string, pubkey rsa.PublicKey) {
	c.rsakeyMap[rkid] = pubkey
	c.lastUpdate = time.Now().Unix()
}

func (c *certsCache) Expired() bool {
	return time.Now().Unix()-c.lastUpdate > 24*60*60
}

func (c *certsCache) CanUpdate() bool {
	return time.Now().Unix()-c.lastUpdate > 60*60
}

var (
	certsCacheInst *certsCache
)

func certsCacheInstance() *certsCache {
	if certsCacheInst == nil {
		certsCacheInst = &certsCache{
			rsakeyMap: make(map[string]rsa.PublicKey, 0),
		}
	}
	return certsCacheInst
}

func reqireCerts(gateway *iggreq.GatewayRequest, gameid string) (ex IGGError) {
	query := iggreq.NewQueryMaker("")
	query.AddParam("game_id", gameid)

	s, e, h := gateway.Get(iggreq.TokenCertsAPIPath, query)
	if !e.IsSuccess() {
		ex = minimumVersionErrorToSDKExceptton(memberVerify, e, h)
		return
	}
	traceid := e.TraceID
	d, e := iggreq.ParseResultFromAPI(s)
	e.TraceID = traceid
	if !e.IsSuccess() {
		if e.Code == 1004 {
			ex = createCGIError(AKApiParam, e.Msg)
			return
		}
		ex = minimumVersionErrorToSDKExceptton(memberVerify, e, h)
		return
	}

	switch o := d.(type) {
	case map[string]interface{}:
		for k, v := range o {
			certsCacheInstance().Add(k, ParseRsaKey(v))
		}
	default:
	}
	return
}

func ParseRsaKey(item interface{}) (ret rsa.PublicKey) {
	switch o := item.(type) {
	case map[string]interface{}:
		buf, _ := json.Marshal(o)
		var jk jwk
		json.Unmarshal(buf, &jk)

		tmp, _ := Base64UrlDecode(jk.N)
		modulus := big.NewInt(0)
		modulus.SetBytes(tmp)
		exponent := big.NewInt(0)
		tmp, _ = Base64UrlDecode(jk.E)
		exponent.SetBytes(tmp)
		ret = rsa.PublicKey{
			N: modulus,
			E: int(exponent.Int64()),
		}

	}
	return
}
