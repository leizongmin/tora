package server

import (
	"fmt"
	"github.com/leizongmin/tora/common"
	"github.com/leizongmin/tora/module/file"
	"github.com/leizongmin/tora/web"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

// 当前版本号
const Version = "1.0"

// X-Powered-By响应头
const PoweredBy = "tora/" + Version

type Server struct {
	Options           Options
	log               *logrus.Logger
	httpServer        *web.Application
	enableModuleFile  bool
	enableModuleShell bool
	enableModuleLog   bool
	moduleFile        *file.ModuleFile
}

type Options struct {
	Log         *logrus.Logger  // 日志输出实例
	Addr        string          // 监听地址，格式：指定端口=:12345 指定地址和端口=127.0.0.1:12345 监听unix-socket=/path/to/sock
	Enable      []string        // 开启的模块，可选：file, shell, log
	FileOptions file.ModuleFile // 文件服务的根目录，如果开启了file模块，需要设置此项
}

type FileOptions = file.ModuleFile

// 默认监听地址
const DefaultListenAddr = ":12345"

func NewServer(options Options) (*Server, error) {
	if len(options.Addr) < 1 {
		options.Addr = DefaultListenAddr
	}

	s := &Server{}

	s.log = options.Log
	if s.log == nil {
		s.log = logrus.New()
		s.log.SetLevel(logrus.ErrorLevel)
	}

	s.httpServer = web.NewApplication()
	s.httpServer.Use(func(ctx *web.Context) {
		s.handleRequest(ctx)
	})

	if len(options.Enable) > 0 {
		for _, n := range options.Enable {
			switch n {
			case "file":
				s.enableModuleFile = true
			case "shell":
				s.enableModuleShell = true
			case "log":
				s.enableModuleLog = true
			default:
				return nil, fmt.Errorf("unsupported module type [%s]", n)
			}
		}
	}

	if s.enableModuleFile {
		if len(options.FileOptions.Root) < 1 {
			return nil, fmt.Errorf("missing option [Root] when module type [file] is enable")
		}
		root, err := filepath.Abs(options.FileOptions.Root)
		if err != nil {
			return nil, err
		}
		options.FileOptions.Log = s.log
		options.FileOptions.Root = root
		s.moduleFile = &options.FileOptions
		s.log.Infof("enable module [file] root=%s", root)
	}

	s.Options = options
	return s, nil
}

func (s *Server) Start() error {
	s.log.Infof("%s listening on %s", PoweredBy, s.Options.Addr)
	return s.httpServer.Listen(s.Options.Addr)
}

func (s *Server) Close() error {
	s.log.Info("trying to close server...")
	return s.httpServer.Close()
}

func (s *Server) handleRequest(ctx *web.Context) {
	ctx.Res.Header().Set("x-powered-by", PoweredBy)
	module := strings.ToLower(ctx.Req.Header.Get("x-module"))

	ctx.Log = s.log.WithFields(logrus.Fields{
		"remote": ctx.Req.RemoteAddr,
		"method": ctx.Req.Method,
		"url":    ctx.Req.RequestURI,
		"module": module,
	})

	switch module {
	case "file":
		s.handleModuleFile(ctx)
	case "shell":
		s.handleModuleShell(ctx)
	case "log":
		s.handleModuleLog(ctx)
	default:
		s.handleModuleError(ctx, module)
	}
}

func (s *Server) handleModuleFile(ctx *web.Context) {
	if !s.enableModuleFile {
		common.ResponseApiError(ctx, "currently not enable [file] module", nil)
		return
	}
	s.moduleFile.Handle(ctx)
}

func (s *Server) handleModuleShell(ctx *web.Context) {
	if !s.enableModuleShell {
		common.ResponseApiError(ctx, "currently not enable [shell] module", nil)
		return
	}
	common.ResponseApiError(ctx, "currently not supported [shell] module", nil)
}

func (s *Server) handleModuleLog(ctx *web.Context) {
	if !s.enableModuleLog {
		common.ResponseApiError(ctx, "currently not enable [log] module", nil)
		return
	}
	common.ResponseApiError(ctx, "currently not supported [log] module", nil)
}

func (s *Server) handleModuleError(ctx *web.Context, name string) {
	if name == "" {
		common.ResponseApiError(ctx, "missing [x-module] header", nil)
		return
	}
	common.ResponseApiError(ctx, fmt.Sprintf("not supported module [%s]", name), nil)
}
