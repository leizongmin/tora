package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s, err := NewServer(Options{Addr: ":12345"})
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, s)
	go s.Start()
	time.Sleep(time.Second)

	{
		req, err := http.NewRequest("GET", "http://127.0.0.1:12345", nil)
		assert.Equal(t, nil, err)
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
		req, err := http.NewRequest("GET", "http://127.0.0.1:12345", nil)
		assert.Equal(t, nil, err)
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
		req, err := http.NewRequest("GET", "http://127.0.0.1:12345", nil)
		assert.Equal(t, nil, err)
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

	s.Close()
}

func TestModuleFile(t *testing.T) {
	name := fmt.Sprintf("tora-%d-%d", time.Now().Unix(), rand.Uint32())
	root := filepath.Join(os.TempDir(), name)
	err := os.Mkdir(root, 0755)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(root)
	file1 := filepath.Join(root, "file1.txt")
	file2 := filepath.Join(root, "file2.txt")
	file1Content := []byte("hello")
	file2Content := []byte("world")
	var file1Stat, file2Stat os.FileInfo
	{
		// 在目录下新建测试用的文件
		if err := ioutil.WriteFile(file1, file1Content, 0666); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(file2, file2Content, 0666); err != nil {
			panic(err)
		}
		if file1Stat, err = os.Stat(file1); err != nil {
			panic(err)
		}
		if file2Stat, err = os.Stat(file2); err != nil {
			panic(err)
		}
	}

	s, err := NewServer(Options{
		Addr:   ":12345",
		Enable: []string{"file"},
		FileOptions: FileOptions{
			FileRoot: root,
			//AllowListDir: true,
			//AllowDelete:  true,
			//AllowPut:     true,
		},
	})
	if err != nil {
		panic(err)
	}
	go s.Start()
	time.Sleep(time.Second)
	{
		// 读取目录，AllowListDir=false 未允许列出目录文件
		req, err := http.NewRequest("GET", "http://127.0.0.1:12345", nil)
		assert.Equal(t, nil, err)
		req.Header.Set("x-module", "file")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		assert.Equal(t, jsonStringify(JSON{
			"ok": true,
			"data": JSON{
				"name":  name,
				"isDir": true,
				"files": nil,
			},
		}), string(body))
	}
	{
		// 设置允许列出目录文件
		s.moduleFile.AllowListDir = true
		// 读取目录
		req, err := http.NewRequest("GET", "http://127.0.0.1:12345", nil)
		assert.Equal(t, nil, err)
		req.Header.Set("x-module", "file")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		data := jsoniter.Get(body)

		assert.Equal(t, true, data.Get("ok").ToBool())
		assert.Equal(t, name, data.Get("data", "name").ToString())
		assert.Equal(t, true, data.Get("data", "isDir").ToBool())
		f1 := jsoniter.Get(body, "data", "files", 0)
		f2 := jsoniter.Get(body, "data", "files", 1)
		assert.Equal(t, 2, jsoniter.Get(body, "data", "files").Size())
		assert.Equal(t, filepath.Base(file1), f1.Get("name").ToString())
		assert.Equal(t, file1Stat.ModTime().String(), f1.Get("modifiedTime").ToString())
		assert.Equal(t, filepath.Base(file2), f2.Get("name").ToString())
		assert.Equal(t, file2Stat.ModTime().String(), f2.Get("modifiedTime").ToString())
	}
	{
		// 获取文件内容
		req, err := http.NewRequest("GET", "http://127.0.0.1:12345/"+filepath.Base(file1), nil)
		assert.Equal(t, nil, err)
		req.Header.Set("x-module", "file")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		assert.Equal(t, file1Content, body)
	}
	file3 := filepath.Join(root, "file3.txt")
	file3Content := []byte("dajdjklfjdksjflkjds")
	file3Md5 := getMd5(file3Content)
	{
		// 上传文件，AllowPut=false 未允许上传
		req, err := http.NewRequest("PUT", "http://127.0.0.1:12345/"+filepath.Base(file3), bytes.NewReader(file3Content))
		assert.Equal(t, nil, err)
		req.Header.Set("x-module", "file")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		data := jsoniter.Get(body)
		assert.Equal(t, false, data.Get("ok").ToBool())
		assert.Equal(t, "not allowed [PUT] file", data.Get("error").ToString())
	}
	{
		// 设置允许上传文件
		s.moduleFile.AllowPut = true
		// 上传文件
		req, err := http.NewRequest("PUT", "http://127.0.0.1:12345/"+filepath.Base(file3), bytes.NewReader(file3Content))
		assert.Equal(t, nil, err)
		req.Header.Set("x-module", "file")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		data := jsoniter.Get(body)
		assert.Equal(t, true, data.Get("ok").ToBool())
		assert.Equal(t, false, data.Get("data", "checkedMd5").ToBool())
	}
	{
		// 上传文件，增加 md5 校验，校验失败
		req, err := http.NewRequest("PUT", "http://127.0.0.1:12345/"+filepath.Base(file3), bytes.NewReader(file3Content))
		assert.Equal(t, nil, err)
		req.Header.Set("x-module", "file")
		req.Header.Set("x-content-md5", "is-must-bad")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		data := jsoniter.Get(body)
		assert.Equal(t, false, data.Get("ok").ToBool())
		assert.Equal(t, fmt.Sprintf("md5 check failed: expected is-must-bad but got %s", file3Md5), data.Get("error").ToString())
	}
	{
		// 上传文件，增加 md5 校验，校验成功
		req, err := http.NewRequest("PUT", "http://127.0.0.1:12345/"+filepath.Base(file3), bytes.NewReader(file3Content))
		assert.Equal(t, nil, err)
		req.Header.Set("x-module", "file")
		req.Header.Set("x-content-md5", strings.ToUpper(file3Md5))
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		data := jsoniter.Get(body)
		assert.Equal(t, true, data.Get("ok").ToBool())
		assert.Equal(t, true, data.Get("data", "checkedMd5").ToBool())
	}
	{
		// 修改文件内容后检查是否正确
		file3Content = []byte(fmt.Sprintf("%s%d", file3Content, rand.Uint32()))
		file3Md5 = getMd5(file3Content)
		{
			req, err := http.NewRequest("PUT", "http://127.0.0.1:12345/"+filepath.Base(file3), bytes.NewReader(file3Content))
			assert.Equal(t, nil, err)
			req.Header.Set("x-module", "file")
			req.Header.Set("x-content-md5", strings.ToUpper(file3Md5))
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			req.WithContext(ctx)
			res, err := http.DefaultClient.Do(req)
			assert.Equal(t, nil, err)
			body, err := ioutil.ReadAll(res.Body)
			assert.Equal(t, nil, err)
			data := jsoniter.Get(body)
			assert.Equal(t, true, data.Get("ok").ToBool())
			assert.Equal(t, true, data.Get("data", "checkedMd5").ToBool())
		}
		{
			// 获取文件内容
			req, err := http.NewRequest("GET", "http://127.0.0.1:12345/"+filepath.Base(file3), nil)
			assert.Equal(t, nil, err)
			req.Header.Set("x-module", "file")
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			req.WithContext(ctx)
			res, err := http.DefaultClient.Do(req)
			assert.Equal(t, nil, err)
			body, err := ioutil.ReadAll(res.Body)
			assert.Equal(t, nil, err)
			assert.Equal(t, file3Content, body)
		}
	}
	{
		// 删除文件，AllowDelete=false 不允许删除
		req, err := http.NewRequest("DELETE", "http://127.0.0.1:12345/"+filepath.Base(file3), bytes.NewReader(file3Content))
		assert.Equal(t, nil, err)
		req.Header.Set("x-module", "file")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Equal(t, nil, err)
		data := jsoniter.Get(body)
		assert.Equal(t, false, data.Get("ok").ToBool())
		assert.Equal(t, "not allowed [DELETE] file", data.Get("error").ToString())
	}
	{
		// 设置允许删除文件
		s.moduleFile.AllowDelete = true
		{
			// 删除文件，成功
			req, err := http.NewRequest("DELETE", "http://127.0.0.1:12345/"+filepath.Base(file3), bytes.NewReader(file3Content))
			assert.Equal(t, nil, err)
			req.Header.Set("x-module", "file")
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			req.WithContext(ctx)
			res, err := http.DefaultClient.Do(req)
			assert.Equal(t, nil, err)
			body, err := ioutil.ReadAll(res.Body)
			assert.Equal(t, nil, err)
			data := jsoniter.Get(body)
			assert.Equal(t, true, data.Get("ok").ToBool())
			assert.Equal(t, true, data.Get("data", "success").ToBool())
		}
		{
			// 获取文件内容，失败
			req, err := http.NewRequest("GET", "http://127.0.0.1:12345/"+filepath.Base(file3), nil)
			assert.Equal(t, nil, err)
			req.Header.Set("x-module", "file")
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			req.WithContext(ctx)
			res, err := http.DefaultClient.Do(req)
			assert.Equal(t, nil, err)
			body, err := ioutil.ReadAll(res.Body)
			assert.Equal(t, 404, res.StatusCode)
			assert.Equal(t, nil, err)
			data := jsoniter.Get(body)
			assert.Equal(t, false, data.Get("ok").ToBool())
		}
	}
	{
		// 判断文件是否存在
		{
			// 不存在
			req, err := http.NewRequest("HEAD", "http://127.0.0.1:12345/abcd/efg", nil)
			assert.Equal(t, nil, err)
			req.Header.Set("x-module", "file")
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			req.WithContext(ctx)
			res, err := http.DefaultClient.Do(req)
			assert.Equal(t, nil, err)
			assert.Equal(t, 500, res.StatusCode)
			assert.Equal(t, "false", res.Header.Get("x-ok"))
			assert.Equal(t, true, len(res.Header.Get("x-error")) > 0)
		}
		{
			// 文件存在
			req, err := http.NewRequest("HEAD", "http://127.0.0.1:12345/"+filepath.Base(file1), nil)
			assert.Equal(t, nil, err)
			req.Header.Set("x-module", "file")
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			req.WithContext(ctx)
			res, err := http.DefaultClient.Do(req)
			assert.Equal(t, nil, err)
			assert.Equal(t, 200, res.StatusCode)
			assert.Equal(t, "true", res.Header.Get("x-ok"))
			assert.Equal(t, "file", res.Header.Get("x-file-type"))
			assert.Equal(t, string(len(file1Content)), res.Header.Get("x-file-size"))
			assert.Equal(t, file1Stat.ModTime().UTC().String(), res.Header.Get("x-last-modified"))
		}
		{
			// 目录存在
			req, err := http.NewRequest("HEAD", "http://127.0.0.1:12345", nil)
			assert.Equal(t, nil, err)
			req.Header.Set("x-module", "file")
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			req.WithContext(ctx)
			res, err := http.DefaultClient.Do(req)
			assert.Equal(t, nil, err)
			assert.Equal(t, 200, res.StatusCode)
			assert.Equal(t, "true", res.Header.Get("x-ok"))
			assert.Equal(t, "dir", res.Header.Get("x-file-type"))
		}
	}
	s.Close()
}

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
