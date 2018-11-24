package server

import (
	"fmt"
	"github.com/leizongmin/tora/common"
	"github.com/leizongmin/tora/module/file"
	"github.com/leizongmin/tora/module/shell"
	"github.com/leizongmin/tora/web"
	"github.com/ryanuber/go-glob"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

// 当前版本号
const Version = "1.0"

// X-Powered-By响应头
const PoweredBy = "tora/" + Version

// 默认监听地址
const DefaultListenAddr = ":12345"

type Server struct {
	Options           Options
	log               *logrus.Logger
	httpServer        *web.Application
	enableModuleFile  bool
	enableModuleShell bool
	enableModuleLog   bool
	moduleFile        *file.ModuleFile
	moduleShell       *shell.ModuleShell
}

type Options struct {
	Log          *logrus.Logger    // 日志输出实例
	Addr         string            // 监听地址，格式：指定端口=:12345 指定地址和端口=127.0.0.1:12345 监听unix-socket=/path/to/sock
	Enable       []string          // 开启的模块，可选：file, shell, log
	FileOptions  file.ModuleFile   // 文件服务配置，如果开启了file模块，需要设置此项
	ShellOptions shell.ModuleShell // 执行命令服务配置，如果开启了shell模块，需要设置此项
	Auth         Auth              // 授权信息
}

type FileOptions = file.ModuleFile
type ShellOptions = shell.ModuleShell

type Auth struct {
	Token     map[string]AuthItem // 允许指定token
	IP        map[string]AuthItem // 允许指定ip
	TokenList []string            // token列表
	IPList    []string            // ip列表
}

type AuthItem struct {
	Allow   bool     // 是否允许访问
	Modules []string // 允许访问的模块
}

type AuthInfo struct {
	Type  string // 类型，如 token 或者 ip
	Token string // 对应的 token
	Ip    string // 对应的 ip
	AuthItem
}

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
		if !(options.FileOptions.DirPerm > 0) {
			options.FileOptions.DirPerm = file.DefaultDirPerm
		}
		if !(options.FileOptions.FilePerm > 0) {
			options.FileOptions.FilePerm = file.DefaultFilePerm
		}
		s.log.Infof("enable module [file] root=%s perm=[dir:%d, file:%d]", root, options.FileOptions.DirPerm, options.FileOptions.FilePerm)
	}

	if s.enableModuleShell {
		if len(options.ShellOptions.Root) < 1 {
			return nil, fmt.Errorf("missing option [Root] when module type [shell] is enable")
		}
		root, err := filepath.Abs(options.ShellOptions.Root)
		if err != nil {
			return nil, err
		}
		options.ShellOptions.Log = s.log
		options.ShellOptions.Root = root
		s.moduleShell = &options.ShellOptions
		if options.ShellOptions.AllowExternalCommands == nil {
			options.ShellOptions.AllowExternalCommands = shell.DefaultAllowExternalCommands
		}
		if options.ShellOptions.AllowInternalCommands == nil {
			options.ShellOptions.AllowInternalCommands = shell.DefaultAllowInternalCommands
		}
		s.log.Infof("enable module [shell] root=%s internalCommands=%s externalCommands=%s", root, options.ShellOptions.AllowInternalCommands, options.ShellOptions.AllowExternalCommands)
	}

	options.Auth.TokenList = make([]string, 0)
	for k, _ := range options.Auth.Token {
		options.Auth.TokenList = append(options.Auth.TokenList, k)
	}
	options.Auth.IPList = make([]string, 0)
	for k, _ := range options.Auth.IP {
		options.Auth.IPList = append(options.Auth.IPList, k)
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

	// 如果x-module=watchdog表示内部健康检查
	if module == "watchdog" {
		common.ResponseApiOk(ctx, common.JSON{"watchdog": true})
		return
	}

	// 检查授权
	auth, ok := s.checkAuth(ctx)
	ctx.Log = ctx.Log.WithFields(logrus.Fields{
		"auth-ok":      ok,
		"auth-type":    auth.Type,
		"auth-ip":      auth.Ip,
		"auth-token":   desensitizeToken(auth.Token),
		"auth-allow":   auth.Allow,
		"auth-modules": strings.Join(auth.Modules, ","),
	})
	if !ok {
		common.ResponseApiErrorWithStatusCode(ctx, 403, "permission denied", nil)
		return
	}

	// 处理请求
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

func (s *Server) checkAuth(ctx *web.Context) (info AuthInfo, ok bool) {
	token := ctx.Req.Header.Get("x-token")
	ip := getIpFromAddr(ctx.Req.RemoteAddr)

	info, ok = s.checkToken(ctx, token)
	if ok {
		return info, true
	}

	info, ok = s.checkIp(ctx, ip)
	if ok {
		return info, true
	}

	info.Ip = ip
	info.Token = token
	return info, false
}

func (s *Server) checkToken(ctx *web.Context, token string) (AuthInfo, bool) {
	a, ok := s.Options.Auth.Token[token]
	info := AuthInfo{Type: "token", Token: token}
	if ok {
		info.Allow = a.Allow
		info.Modules = a.Modules
	} else {
		// 如果无法直接匹配，则尝试通配模式
		for _, v := range s.Options.Auth.TokenList {
			if glob.Glob(v, token) {
				a, _ = s.Options.Auth.Token[v]
				info = AuthInfo{Type: "token", Token: v}
				info.Allow = a.Allow
				info.Modules = a.Modules
				ok = true
				break
			}
		}
	}
	return info, ok
}

func (s *Server) checkIp(ctx *web.Context, ip string) (AuthInfo, bool) {
	a, ok := s.Options.Auth.IP[ip]
	info := AuthInfo{Type: "ip", Ip: ip}
	if ok {
		info.Allow = a.Allow
		info.Modules = a.Modules
	} else {
		// 如果无法直接匹配，则尝试通配模式
		for _, v := range s.Options.Auth.IPList {
			if glob.Glob(v, ip) {
				a, _ = s.Options.Auth.IP[v]
				info = AuthInfo{Type: "token", Token: v}
				info.Allow = a.Allow
				info.Modules = a.Modules
				ok = true
				break
			}
		}
	}
	return info, ok
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

func getIpFromAddr(addr string) string {
	if strings.Contains(addr, "/") {
		return addr
	}
	s := strings.Split(addr, ":")
	return s[0]
}

func desensitizeToken(token string) string {
	size := len(token)
	if size == 0 {
		return ""
	}
	if size == 1 {
		return "*"
	}
	if size == 2 {
		return token[0:1] + "*"
	}
	if size < 4 {
		return token[0:2] + "****"
	}
	return token[0:2] + "****" + token[len(token)-2:]
}
