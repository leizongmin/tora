package file

import (
	"fmt"
	"github.com/leizongmin/tora/common"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ModuleFile struct {
	Log          *logrus.Logger // 日志模块
	Root         string         // 文件根目录
	AllowPut     bool           // 允许上传文件
	AllowDelete  bool           // 允许删除文件
	AllowListDir bool           // 允许列出目录
}

func (m *ModuleFile) Handle(log *logrus.Entry, w http.ResponseWriter, r *http.Request) {
	f, err := resolveFilePath(m.Root, r.URL.Path)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
	}
	switch r.Method {
	case "HEAD":
		m.handleHead(log, w, r, f)
	case "GET":
		m.handleGet(log, w, r, f)
	case "PUT":
		m.handlePut(log, w, r, f)
	case "DELETE":
		m.handleDelete(log, w, r, f)
	default:
		common.ResponseApiError(log, w, fmt.Sprintf("method [%s] not allowed", r.Method), nil)
	}
}

func (m *ModuleFile) handleHead(log *logrus.Entry, w http.ResponseWriter, r *http.Request, f string) {
	s, err := os.Stat(f)
	if err != nil {
		w.Header().Set("x-ok", "false")
		w.Header().Set("x-error", err.Error())
		w.WriteHeader(500)
		return
	}
	w.Header().Set("x-ok", "true")
	if s.IsDir() {
		w.Header().Set("x-file-type", "dir")
	} else {
		w.Header().Set("x-file-type", "file")
		w.Header().Set("x-file-size", string(s.Size()))
		w.Header().Set("x-last-modified", s.ModTime().UTC().String())
	}
}

func (m *ModuleFile) handleGet(log *logrus.Entry, w http.ResponseWriter, r *http.Request, f string) {
	s, err := os.Stat(f)
	if err != nil {
		common.ResponseApiErrorWithStatusCode(log, w, 404, err.Error(), nil)
		return
	}
	if s.IsDir() {
		if m.AllowListDir {
			m.responseDirList(log, w, f, s)
		} else {
			m.responseDirInfo(log, w, f, s)
		}
		return
	}
	m.responseFileContent(log, w, f, s)
}

func (m *ModuleFile) handlePut(log *logrus.Entry, w http.ResponseWriter, r *http.Request, f string) {
	if !m.AllowPut {
		common.ResponseApiError(log, w, "not allowed [PUT] file", nil)
		return
	}

	md5 := r.Header.Get("x-content-md5")
	dir := filepath.Dir(f)
	tmpFile := filepath.Join(dir, fmt.Sprintf(".%s.%d-%d", filepath.Base(f), time.Now().Unix(), rand.Uint32()))

	// 先保证目录存在
	if err := os.MkdirAll(dir, 0766); err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
		return
	}

	// 先存储到临时文件
	tmpFd, err := os.Create(tmpFile)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
		return
	}
	defer tmpFd.Close()
	_, err = io.Copy(tmpFd, r.Body)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
		return
	}

	// 校验md5值
	checkedMd5 := false
	if len(md5) > 0 {
		tmpMd5, err := getFileMd5(tmpFile)
		if err != nil {
			common.ResponseApiError(log, w, err.Error(), nil)
			return
		}
		if strings.ToLower(tmpMd5) != strings.ToLower(md5) {
			common.ResponseApiError(log, w, fmt.Sprintf("md5 check failed: expected %s but got %s", md5, tmpMd5), common.JSON{"expected": md5, "actual": tmpMd5})
			return
		}
		checkedMd5 = true
	}

	// 删除旧文件，覆盖新文件
	err = os.Remove(f)
	if err != nil && !os.IsNotExist(err) {
		common.ResponseApiError(log, w, err.Error(), nil)
		return
	}
	err = os.Rename(tmpFile, f)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
		return
	}

	common.ResponseApiOk(log, w, common.JSON{"checkedMd5": checkedMd5})
}

func (m *ModuleFile) handleDelete(log *logrus.Entry, w http.ResponseWriter, r *http.Request, f string) {
	if !m.AllowDelete {
		common.ResponseApiError(log, w, "not allowed [DELETE] file", nil)
		return
	}

	s, err := os.Stat(f)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
	}
	if s.IsDir() {
		m.responseDeleteDir(log, w, f, s)
		return
	}
	m.responseDeleteFile(log, w, f, s)
}

func (m *ModuleFile) responseDirList(log *logrus.Entry, w http.ResponseWriter, f string, s os.FileInfo) {
	list, err := ioutil.ReadDir(f)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
		return
	}
	list2 := make([]common.JSON, len(list))
	for i, v := range list {
		list2[i] = common.JSON{
			"name":         v.Name(),
			"size":         v.Size(),
			"modifiedTime": v.ModTime().String(),
		}
	}
	common.ResponseApiOk(log, w, common.JSON{
		"name":  s.Name(),
		"isDir": true,
		"files": list2,
	})
}

func (m *ModuleFile) responseDirInfo(log *logrus.Entry, w http.ResponseWriter, f string, s os.FileInfo) {
	common.ResponseApiOk(log, w, common.JSON{
		"name":  s.Name(),
		"isDir": true,
		"files": nil,
	})
}

func (m *ModuleFile) responseFileContent(log *logrus.Entry, w http.ResponseWriter, f string, s os.FileInfo) {
	r, err := os.Open(f)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
		return
	}
	defer r.Close()
	w.Header().Set("content-type", "application/octet-stream")
	w.Header().Set("x-file-size", string(s.Size()))
	w.Header().Set("x-last-modified", s.ModTime().UTC().String())
	_, err = io.Copy(w, r)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
	}
}

func (m *ModuleFile) responseDeleteDir(log *logrus.Entry, w http.ResponseWriter, f string, s os.FileInfo) {
	err := os.RemoveAll(f)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
		return
	}
	common.ResponseApiOk(log, w, common.JSON{"success": true})
}

func (m *ModuleFile) responseDeleteFile(log *logrus.Entry, w http.ResponseWriter, f string, s os.FileInfo) {
	err := os.Remove(f)
	if err != nil {
		common.ResponseApiError(log, w, err.Error(), nil)
		return
	}
	common.ResponseApiOk(log, w, common.JSON{"success": true})
}
