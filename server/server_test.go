package server

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s := NewServer(Options{Addr: ":12345"})
	assert.NotEqual(t, nil, s)
	go s.Start()
	time.Sleep(time.Second)

	req, err := http.NewRequest("GET", "http://127.0.0.1:12345", nil)
	assert.Equal(t, nil, err)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	assert.Equal(t, nil, err)
	body, err := ioutil.ReadAll(res.Body)
	assert.Equal(t, nil, err)
	assert.Equal(t, "{\"data\":null,\"error\":\"missing x-module header\",\"ok\":false}", string(body))
	fmt.Println(res.Header)

	s.Close()
}
