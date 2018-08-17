package server

import (
	"fmt"
	"github.com/leizongmin/tora/common"
	"github.com/leizongmin/tora/module/file"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
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
	httpServer        *http.Server
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

	s.httpServer = &http.Server{}
	s.httpServer.Addr = options.Addr
	s.httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handleRequest(w, r)
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
	}

	return s, nil
}

func (s *Server) Start() error {
	var proto string
	if s.httpServer.Addr == "" {
		s.httpServer.Addr = ":http"
	}
	if strings.Contains(s.httpServer.Addr, "/") {
		proto = "unix"
	} else {
		proto = "tcp"
	}
	l, err := net.Listen(proto, s.httpServer.Addr)
	if err != nil {
		return err
	}
	s.log.Infof("%s listening on %s", PoweredBy, s.httpServer.Addr)
	return s.httpServer.Serve(l)
}

func (s *Server) Close() error {
	s.log.Info("trying to close server...")
	return s.httpServer.Close()
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("x-powered-by", PoweredBy)
	module := strings.ToLower(r.Header.Get("x-module"))

	log := s.log.WithFields(logrus.Fields{
		"remote": r.RemoteAddr,
		"method": r.Method,
		"url":    r.RequestURI,
		"module": module,
	})

	switch module {
	case "file":
		s.handleModuleFile(log, w, r)
	case "shell":
		s.handleModuleShell(log, w, r)
	case "log":
		s.handleModuleLog(log, w, r)
	default:
		s.handleModuleError(log, w, r, module)
	}
}

func (s *Server) handleModuleFile(log *logrus.Entry, w http.ResponseWriter, r *http.Request) {
	if !s.enableModuleFile {
		common.ResponseApiError(log, w, "currently not enable [file] module", nil)
		return
	}
	s.moduleFile.Handle(log, w, r)
}

func (s *Server) handleModuleShell(log *logrus.Entry, w http.ResponseWriter, r *http.Request) {
	if !s.enableModuleShell {
		common.ResponseApiError(log, w, "currently not enable [shell] module", nil)
		return
	}
	common.ResponseApiError(log, w, "currently not supported [shell] module", nil)
}

func (s *Server) handleModuleLog(log *logrus.Entry, w http.ResponseWriter, r *http.Request) {
	if !s.enableModuleLog {
		common.ResponseApiError(log, w, "currently not enable [log] module", nil)
		return
	}
	common.ResponseApiError(log, w, "currently not supported [log] module", nil)
}

func (s *Server) handleModuleError(log *logrus.Entry, w http.ResponseWriter, r *http.Request, name string) {
	if name == "" {
		common.ResponseApiError(log, w, "missing [x-module] header", nil)
		return
	}
	common.ResponseApiError(log, w, fmt.Sprintf("not supported module [%s]", name), nil)
}
