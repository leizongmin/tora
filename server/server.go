package server

import (
	"fmt"
	"github.com/leizongmin/tora/common"
	"github.com/leizongmin/tora/module/file"
	"net"
	"net/http"
	"path/filepath"
	"strings"
)

type Server struct {
	Options           Options
	httpServer        *http.Server
	enableModuleFile  bool
	enableModuleShell bool
	enableModuleLog   bool
	moduleFile        *file.ModuleFile
}

type Options struct {
	Addr     string   `json:"addr"`     // 监听地址，格式：指定端口=:12345 指定地址和端口=127.0.0.1:12345 监听unix-socket=/path/to/sock
	Enable   []string `json:"enable"`   // 开启的模块，可选：file, shell, log
	FileRoot string   `json:"fileRoot"` // 文件服务的根目录，如果开启了file模块，需要设置此项
}

// 默认监听地址
const DefaultListenAddr = ":12345"

func NewServer(options Options) (*Server, error) {
	if len(options.Addr) < 1 {
		options.Addr = DefaultListenAddr
	}

	s := &Server{}
	s.httpServer = &http.Server{}
	s.httpServer.Addr = options.Addr
	s.httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("x-powered-by", "tora/1.0")
		module := strings.ToLower(r.Header.Get("x-module"))
		switch module {
		case "file":
			s.handleModuleFile(w, r)
		case "shell":
			s.handleModuleShell(w, r)
		case "log":
			s.handleModuleLog(w, r)
		default:
			s.handleModuleError(w, r, module)
		}
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
		if len(options.FileRoot) < 1 {
			return nil, fmt.Errorf("missing option [FileRoot] when module type [file] is enable")
		}
		root, err := filepath.Abs(options.FileRoot)
		if err != nil {
			return nil, err
		}
		s.moduleFile = &file.ModuleFile{FileRoot: root}
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
	return s.httpServer.Serve(l)
}

func (s *Server) Close() error {
	return s.httpServer.Close()
}

func (s *Server) handleModuleFile(w http.ResponseWriter, r *http.Request) {
	if !s.enableModuleFile {
		common.ResponseApiError(w, "currently not enable [file] module", nil)
		return
	}
	s.moduleFile.Handle(w, r)
}

func (s *Server) handleModuleShell(w http.ResponseWriter, r *http.Request) {
	if !s.enableModuleShell {
		common.ResponseApiError(w, "currently not enable [shell] module", nil)
		return
	}
	common.ResponseApiError(w, "currently not supported [shell] module", nil)
}

func (s *Server) handleModuleLog(w http.ResponseWriter, r *http.Request) {
	if !s.enableModuleLog {
		common.ResponseApiError(w, "currently not enable [log] module", nil)
		return
	}
	common.ResponseApiError(w, "currently not supported [log] module", nil)
}

func (s *Server) handleModuleError(w http.ResponseWriter, r *http.Request, name string) {
	if name == "" {
		common.ResponseApiError(w, "missing [x-module] header", nil)
		return
	}
	common.ResponseApiError(w, fmt.Sprintf("not supported module [%s]", name), nil)
}
