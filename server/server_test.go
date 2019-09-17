package server

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	addr, url := getRandomPort()
	s, err := NewServer(Options{
		Addr: addr,
		Auth: Auth{
			Token: map[string]AuthItem{
				"testtoken": {
					Allow:   true,
					Modules: []string{"file"},
				},
				"super*": {
					Allow:   true,
					Modules: []string{"file"},
				},
			},
		},
	})
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, s)
	go s.Start()
	time.Sleep(time.Second)

	{
		req, err := http.NewRequest("GET", url, nil)
		assert.Equal(t, nil, err)
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		assert.Equal(t, jsonStringify(JSON{
			"ok":    false,
			"error": "permission denied",
			"data":  nil,
		}), string(body))
	}
	{
		req, err := http.NewRequest("GET", url, nil)
		assert.Equal(t, nil, err)
		req.Header.Set("x-token", "bad")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		assert.Equal(t, jsonStringify(JSON{
			"ok":    false,
			"error": "permission denied",
			"data":  nil,
		}), string(body))
	}
	{
		req, err := http.NewRequest("GET", url, nil)
		assert.Equal(t, nil, err)
		req.Header.Set("x-token", "testtoken")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		assert.Equal(t, jsonStringify(JSON{
			"ok":    false,
			"error": "missing [x-module] header",
			"data":  nil,
		}), string(body))
	}
	{
		req, err := http.NewRequest("GET", url, nil)
		assert.Equal(t, nil, err)
		req.Header.Set("x-token", "testtoken")
		req.Header.Set("x-module", "hello")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		assert.Equal(t, jsonStringify(JSON{
			"ok":    false,
			"error": "not supported module [hello]",
			"data":  nil,
		}), string(body))
	}
	{
		req, err := http.NewRequest("GET", url, nil)
		assert.Equal(t, nil, err)
		req.Header.Set("x-token", "testtoken")
		req.Header.Set("x-module", "file")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		assert.Equal(t, jsonStringify(JSON{
			"ok":    false,
			"error": "currently not enable [file] module",
			"data":  nil,
		}), string(body))
	}
	{
		// 通配符模式
		req, err := http.NewRequest("GET", url, nil)
		assert.Equal(t, nil, err)
		req.Header.Set("x-token", "super666")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		assert.Equal(t, jsonStringify(JSON{
			"ok":    false,
			"error": "missing [x-module] header",
			"data":  nil,
		}), string(body))
	}
	{
		// 通配符模式
		req, err := http.NewRequest("GET", url, nil)
		assert.Equal(t, nil, err)
		req.Header.Set("x-token", "super 789")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		assert.Equal(t, jsonStringify(JSON{
			"ok":    false,
			"error": "missing [x-module] header",
			"data":  nil,
		}), string(body))
	}

	s.Close()
}
