package server

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

type JSON map[string]interface{}

func jsonStringify(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func getMd5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:16])
}
