package file

import (
	"fmt"
	"github.com/leizongmin/tora/common"
	"github.com/leizongmin/tora/web"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 默认创建目录权限
const DefaultDirPerm = 0777

// 默认创建文件权限
const DefaultFilePerm = 0666

type ModuleFile struct {
	Log          *logrus.Logger // 日志模块
	Root         string         // 文件根目录
	AllowPut     bool           // 允许上传文件
	AllowDelete  bool           // 允许删除文件
	AllowListDir bool           // 允许列出目录
	DirPerm      os.FileMode    // 创建的目录权限
	FilePerm     os.FileMode    // 创建的文件权限
}

func (m *ModuleFile) Handle(ctx *web.Context) {
	f, err := resolveFilePath(m.Root, ctx.Req.URL.Path)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}
	switch ctx.Req.Method {
	case "HEAD":
		m.handleHead(ctx, f)
	case "GET":
		m.handleGet(ctx, f)
	case "PUT":
		m.handlePut(ctx, f)
	case "DELETE":
		m.handleDelete(ctx, f)
	default:
		common.ResponseApiError(ctx, fmt.Sprintf("method [%s] not allowed", ctx.Req.Method), nil)
	}
}

func (m *ModuleFile) handleHead(ctx *web.Context, f string) {
	s, err := os.Stat(f)
	if err != nil {
		ctx.Res.Header().Set("x-ok", "false")
		ctx.Res.Header().Set("x-error", err.Error())
		ctx.Res.WriteHeader(500)
		return
	}
	ctx.Res.Header().Set("x-ok", "true")
	if s.IsDir() {
		ctx.Res.Header().Set("x-file-type", "dir")
	} else {
		ctx.Res.Header().Set("x-file-type", "file")
		ctx.Res.Header().Set("x-file-size", string(s.Size()))
		ctx.Res.Header().Set("x-last-modified", s.ModTime().UTC().String())
	}
}

func (m *ModuleFile) handleGet(ctx *web.Context, f string) {
	s, err := os.Stat(f)
	if err != nil {
		common.ResponseApiErrorWithStatusCode(ctx, 404, err.Error(), nil)
		return
	}
	if s.IsDir() {
		if m.AllowListDir {
			m.responseDirList(ctx, f, s)
		} else {
			m.responseDirInfo(ctx, f, s)
		}
		return
	}
	m.responseFileContent(ctx, f, s)
}

func (m *ModuleFile) handlePut(ctx *web.Context, f string) {
	if !m.AllowPut {
		common.ResponseApiError(ctx, "not allowed [PUT] file", nil)
		return
	}

	md5 := ctx.Req.Header.Get("x-content-md5")
	dir := filepath.Dir(f)
	tmpFile := filepath.Join(dir, fmt.Sprintf(".%s.%d-%d", filepath.Base(f), time.Now().Unix(), rand.Uint32()))

	// 先保证目录存在
	if err := os.MkdirAll(dir, m.DirPerm); err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}

	// 先存储到临时文件
	tmpFd, err := os.Create(tmpFile)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}
	defer tmpFd.Close()
	_, err = io.Copy(tmpFd, ctx.Req.Body)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}

	// 校验md5值
	checkedMd5 := false
	if len(md5) > 0 {
		tmpMd5, err := getFileMd5(tmpFile)
		if err != nil {
			common.ResponseApiError(ctx, err.Error(), nil)
			return
		}
		if strings.ToLower(tmpMd5) != strings.ToLower(md5) {
			common.ResponseApiError(ctx, fmt.Sprintf("md5 check failed: expected %s but got %s", md5, tmpMd5), common.JSON{"expected": md5, "actual": tmpMd5})
			return
		}
		checkedMd5 = true
	}

	// 删除旧文件，覆盖新文件
	err = os.Remove(f)
	if err != nil && !os.IsNotExist(err) {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}
	err = os.Rename(tmpFile, f)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}

	// 更改文件权限
	err = os.Chmod(f, m.FilePerm)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}

	common.ResponseApiOk(ctx, common.JSON{"checkedMd5": checkedMd5})
}

func (m *ModuleFile) handleDelete(ctx *web.Context, f string) {
	if !m.AllowDelete {
		common.ResponseApiError(ctx, "not allowed [DELETE] file", nil)
		return
	}

	s, err := os.Stat(f)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}
	if s.IsDir() {
		m.responseDeleteDir(ctx, f, s)
		return
	}
	m.responseDeleteFile(ctx, f, s)
}

func (m *ModuleFile) responseDirList(ctx *web.Context, f string, s os.FileInfo) {
	list, err := ioutil.ReadDir(f)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}
	list2 := make([]common.JSON, len(list))
	for i, v := range list {
		list2[i] = common.JSON{
			"name":         v.Name(),
			"isDir":        v.IsDir(),
			"size":         v.Size(),
			"modifiedTime": v.ModTime().String(),
		}
	}
	ctx.Res.Header().Set("x-file-type", "dir")
	common.ResponseApiOk(ctx, common.JSON{
		"name":  s.Name(),
		"isDir": true,
		"files": list2,
	})
}

func (m *ModuleFile) responseDirInfo(ctx *web.Context, f string, s os.FileInfo) {
	ctx.Res.Header().Set("x-file-type", "dir")
	common.ResponseApiOk(ctx, common.JSON{
		"name":  s.Name(),
		"isDir": true,
		"files": nil,
	})
}

func (m *ModuleFile) responseFileContent(ctx *web.Context, f string, s os.FileInfo) {
	r, err := os.Open(f)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}
	defer r.Close()
	ctx.Res.Header().Set("x-file-type", "file")
	ctx.Res.Header().Set("content-type", "application/octet-stream")
	ctx.Res.Header().Set("x-file-size", string(s.Size()))
	ctx.Res.Header().Set("x-last-modified", s.ModTime().UTC().String())
	_, err = io.Copy(ctx.Res, r)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
	}
}

func (m *ModuleFile) responseDeleteDir(ctx *web.Context, f string, s os.FileInfo) {
	err := os.RemoveAll(f)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}
	common.ResponseApiOk(ctx, common.JSON{"success": true})
}

func (m *ModuleFile) responseDeleteFile(ctx *web.Context, f string, s os.FileInfo) {
	err := os.Remove(f)
	if err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}
	common.ResponseApiOk(ctx, common.JSON{"success": true})
}
