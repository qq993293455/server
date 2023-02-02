package apis

import (
	"crypto/aes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

type accessTokenHeader struct {
	Akid   string `json:"akid,int64"`
	Rkid   string `json:"rkid,int64"`
	Gameid string `json:"g,int64"`
}

type encodedToken struct {
	Header    string
	Body      string
	Signature string
}

func (et *encodedToken) DecodHeader() (accessTokenHeader, error) {
	ret, _ := Base64UrlDecode(et.Header)
	return parseHeader(ret)
}

func (et *encodedToken) DecodeBody(aesKey string) (token AccessToken, err error) {
	ciphertext, _ := Base64UrlDecode(et.Body)
	aesCipher, _ := aes.NewCipher([]byte(aesKey))

	decrypted := make([]byte, len(ciphertext))
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()
	size := aes.BlockSize
	if len(ciphertext) < aes.BlockSize {
		size = len(ciphertext)
	}
	bs, be := 0, size
	for bs < len(ciphertext) {
		aesCipher.Decrypt(decrypted[bs:be], ciphertext[bs:be])
		bs = bs + size
		if len(ciphertext)-bs < aes.BlockSize {
			size = len(ciphertext) - bs
		}
		be = be + size
	}
	decrypted = pkcs7Unpad(decrypted)
	return parseBody(decrypted)
}

func (et *encodedToken) DecodSignature() []byte {
	ret, _ := Base64UrlDecode(et.Signature)
	modulus := big.NewInt(0)
	modulus.SetBytes(ret)

	return ret
}

func parseHeader(tokenHeader []byte) (ret accessTokenHeader, err error) {
	var jsonObj map[string]interface{}
	err = json.Unmarshal(tokenHeader, &jsonObj)
	if err == nil {
		ret.Akid = GetJosnValueAsString(jsonObj, "akid")
		ret.Rkid = GetJosnValueAsString(jsonObj, "rkid")
		ret.Gameid = GetJosnValueAsString(jsonObj, "g")
	}
	return
}

func parseBody(tokenBody []byte) (ret AccessToken, err error) {

	var jsonObj map[string]interface{}
	err = json.Unmarshal(tokenBody, &jsonObj)
	if err == nil {
		ret.Iggid = GetJosnValueAsString(jsonObj, "sub")
		ret.LoginType = GetJosnValueAsString(jsonObj, "typ")
		ret.HashedUDID = GetJosnValueAsString(jsonObj, "udid")
		ret.Ttl = GetJosnValueAsInt64(jsonObj, "ttl")
		ret.Iat = GetJosnValueAsInt64(jsonObj, "iat")
	}
	return
}

func pkcs7Unpad(data []byte) []byte {
	n := int(data[len(data)-1])
	if n > 0 && n < aes.BlockSize {
		for _, c := range data[len(data)-n:] {
			if int(c) != n {
				n = 0
				break
			}
		}
		return data[:len(data)-n]
	}
	return data
}

func GetJosnValueAsString(jsonObj map[string]interface{}, name string) string {
	if value, ok := jsonObj[name]; ok {
		switch o := value.(type) {
		case float64:
			return strconv.FormatInt(int64(o), 10)
		case string:
			return o
		default:
		}
	}
	return ""
}

func GetJosnValueAsInt64(jsonObj map[string]interface{}, name string) int64 {
	if value, ok := jsonObj[name]; ok {
		switch o := value.(type) {
		case float64:
			return int64(o)
		default:
		}
	}
	return 0
}
