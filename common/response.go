package common

import (
	"encoding/json"
	"fmt"
	"github.com/leizongmin/tora/web"
	"strings"
)

func ResponseJson(ctx *web.Context, statusCode int, data JSON) {
	b, err := json.Marshal(data)
	if err != nil {
		ctx.Res.WriteHeader(500)
		msg := strings.Replace(err.Error(), "\"", "\\\"", -1)
		ctx.Res.Write([]byte(fmt.Sprintf("{\"ok\":false,\"error\":\"JSON stringify failed: %s\"", msg)))
		return
	}
	ctx.Res.Header().Set("content-type", "application/json")
	ctx.Res.WriteHeader(statusCode)
	ctx.Res.Write(b)
}

func ResponseApiOk(ctx *web.Context, data JSON) {
	d := JSON{"ok": true, "data": data}
	ctx.Log.Info("OK")
	ResponseJson(ctx, 200, d)
}

func ResponseApiError(ctx *web.Context, err string, data JSON) {
	d := JSON{"ok": false, "error": err, "data": data}
	ctx.Log.WithField("error", err).Warn("Error")
	ResponseJson(ctx, 500, d)
}

func ResponseApiErrorWithStatusCode(ctx *web.Context, statusCode int, err string, data JSON) {
	d := JSON{"ok": false, "error": err, "data": data}
	ctx.Log.WithField("error", err).Warn("Error")
	ResponseJson(ctx, statusCode, d)
}
