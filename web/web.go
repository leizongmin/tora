package web

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
)

type Application struct {
	server   http.Server
	listener net.Listener
	addr     string
	handlers []HandlerFunc
}

type Context struct {
	handlerIndex int
	app          *Application
	Req          *http.Request
	Res          http.ResponseWriter
	Log          logrus.FieldLogger
}

type HandlerFunc = func(ctx *Context)

func NewApplication() *Application {
	app := Application{}
	app.handlers = make([]HandlerFunc, 0)
	return &app
}

func (app *Application) Listen(addr string) error {
	app.server = http.Server{}
	app.server.Addr = addr
	app.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.handleRequest(w, r)
	})

	var proto string
	if app.server.Addr == "" {
		app.server.Addr = ":http"
	}
	if strings.Contains(app.server.Addr, "/") {
		proto = "unix"
	} else {
		proto = "tcp"
	}
	l, err := net.Listen(proto, app.server.Addr)
	if err != nil {
		return err
	}
	app.listener = l
	return app.server.Serve(l)
}

func (app *Application) Close() error {
	return app.server.Close()
}

func (app *Application) Use(fn HandlerFunc) {
	app.handlers = append(app.handlers, fn)
}

func (app *Application) handleRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	app.execHandler(0, NewContext(w, r))
}

func (app *Application) execHandler(index int, ctx *Context) {
	if index < len(app.handlers) {
		fn := app.handlers[index]
		fn(ctx)
		return
	}
	ctx.Res.WriteHeader(500)
	ctx.Res.Write([]byte(fmt.Sprintf("Handler #%s does not exists", index)))
}

func (ctx *Context) Next() {
	ctx.handlerIndex++
	if ctx.handlerIndex < len(ctx.app.handlers) {
		ctx.app.execHandler(ctx.handlerIndex, ctx)
		return
	}
	ctx.Res.WriteHeader(404)
	ctx.Res.Write([]byte(fmt.Sprintf("Cannot %s %s", ctx.Req.Method, ctx.Req.RequestURI)))
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{Req: r, Res: w}
}
