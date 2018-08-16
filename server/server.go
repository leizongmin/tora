package server

import (
	"fmt"
	"github.com/leizongmin/tora/common"
	"net"
	"net/http"
	"strings"
)

type Server struct {
	Options    Options
	httpServer *http.Server
}

type Options struct {
	Addr string `json:"addr"` // 监听地址，格式：指定端口=:12345 指定地址和端口=127.0.0.1:12345 监听unix-socket=/path/to/sock
}

const DefaultListenAddr = ":12345"

func NewServer(options Options) *Server {
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
	return s
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
	fmt.Println(l)
	return s.httpServer.Serve(l)
}

func (s *Server) Close() error {
	return s.httpServer.Close()
}

func (s *Server) handleModuleFile(w http.ResponseWriter, r *http.Request) {
	common.ResponseApiError(w, "currently not support file module", nil)
}

func (s *Server) handleModuleShell(w http.ResponseWriter, r *http.Request) {
	common.ResponseApiError(w, "currently not support shell module", nil)
}

func (s *Server) handleModuleLog(w http.ResponseWriter, r *http.Request) {
	common.ResponseApiError(w, "currently not support log module", nil)
}

func (s *Server) handleModuleError(w http.ResponseWriter, r *http.Request, name string) {
	if name == "" {
		common.ResponseApiError(w, "missing x-module header", nil)
		return
	}
	common.ResponseApiError(w, fmt.Sprintf("not supported module: %s", name), nil)
}
