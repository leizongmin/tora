package server

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
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

func getRandomPort() (addr string, url string) {
	p := strconv.Itoa(int(rand.Int31n(20000) + 40000))
	addr = ":" + p
	url = "http://127.0.0.1:" + p
	fmt.Println("random addr:", url)
	return addr, url
}
