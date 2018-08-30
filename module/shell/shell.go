package shell

import (
	"fmt"
	"github.com/leizongmin/tora/common"
	"github.com/leizongmin/tora/web"
	"github.com/sirupsen/logrus"
)

type ModuleShell struct {
	Log                   *logrus.Logger // 日志模块
	Root                  string         // 文件根目录
	AllowInternalCommands []string       // 允许执行的内部命令
	AllowExternalCommands []string       // 允许执行的外部命令
}

var DefaultAllowInternalCommands = []string{"list", "cd", "cat", "exit", "run"}
var DefaultAllowExternalCommands = []string{}

type ExecInfo struct {
	CWD       string            `json:"cwd"`
	Define    map[string]string `json:"define"`
	Run       []string          `json:"run"`
	OnSuccess []string          `json:"onSuccess"`
	OnError   []string          `json:"onError"`
	OnEnd     []string          `json:"onEnd"`
}

func (m *ModuleShell) Handle(ctx *web.Context) {
	if ctx.Req.Method != "POST" {
		common.ResponseApiError(ctx, fmt.Sprintf("method [%s] not allowed", ctx.Req.Method), nil)
		return
	}
	info := ExecInfo{}
	if err := ctx.Util.ParseBodyJson(&info); err != nil {
		common.ResponseApiError(ctx, err.Error(), nil)
		return
	}

	ctx.Util.ResponseJson(info)
}
