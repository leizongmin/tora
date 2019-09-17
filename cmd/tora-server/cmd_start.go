package main

import (
	"context"
	"flag"
	"github.com/coreos/go-systemd/daemon"
	"github.com/leizongmin/tora/server"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func cmdStart(args []string) {
	var configFile string
	log := logrus.New()

	// 解析命令行参数
	cmd := flag.NewFlagSet(CmdName, flag.ExitOnError)
	cmd.StringVar(&configFile, "c", DefaultConfigFilePath, "set config file path")
	cmd.Usage = func() {
		printUsage(cmd)
	}
	cmd.Parse(args)

	log.Infof("PID: %d", os.Getpid())

	// 读取配置文件
	c, err := LoadConfigFile(configFile)
	if err != nil {
		log.Fatalf("Load config failed: %s\n", err)
	}

	// 设置日志记录器
	if level, err := logrus.ParseLevel(c.Log.Level); err != nil {
		log.Errorf("Invalid log level: %s\n", c.Log.Level)
	} else {
		log.SetLevel(level)
	}
	switch c.Log.Format {
	case "text":
		log.Formatter = &logrus.TextFormatter{}
	case "json":
		log.Formatter = &logrus.JSONFormatter{}
	default:
		log.Errorf("Invalid log format: %s\n", c.Log.Format)
	}

	// 创建服务器实例
	s, err := server.NewServer(server.Options{
		Log:    log,
		Addr:   c.Listen,
		Enable: c.Enable,
		FileOptions: server.FileOptions{
			Root:         c.Module.File.Root,
			AllowDelete:  c.Module.File.AllowDelete,
			AllowPut:     c.Module.File.AllowPut,
			AllowListDir: c.Module.File.AllowListDir,
		},
		Auth: server.Auth{
			Token: mapConfigAuthItemToServerAuthItem(c.Auth.Token),
			IP:    mapConfigAuthItemToServerAuthItem(c.Auth.IP),
		},
	})
	if err != nil {
		log.Panicf("Try to start server failed: %s\n", err)
	}

	// systemd daemon
	ok, err := daemon.SdNotify(false, "READY=1")
	if err != nil {
		log.Errorf("[daemon] notification supported, but failure happened: %s", err)
	} else if !ok {
		log.Errorln("[daemon] notification not supported")
	} else {
		log.Infoln("[daemon] notification supported, data has been sent")
	}
	// systemd watchdog
	go func() {
		interval, err := daemon.SdWatchdogEnabled(false)
		if err != nil || interval == 0 {
			log.Warnln("[daemon] watchdog is not enabled")
			return
		}
		time.Sleep(interval / 3)
		for {
			//_, err := http.Get("http://" + c.Listen)
			req, err := http.NewRequest("GET", "http://"+c.Listen, nil)
			if err != nil {
				log.Warnf("[daemon] watchdog check failed: %s\n", err)
				continue
			}
			req.Header.Set("x-module", "watchdog")
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			req.WithContext(ctx)
			_, err = http.DefaultClient.Do(req)
			if err == nil {
				daemon.SdNotify(false, "WATCHDOG=1")
			} else {
				log.Warnf("[daemon] watchdog check failed: %s\n", err)
			}
			time.Sleep(interval / 3)
		}
	}()

	// 接收系统信号
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		sig := <-sigs
		log.Warnf("RECEIVED SIGNAL: %s\n", sig)
		s.Close()
		os.Exit(1)
	}()

	// 监听端口
	if err := s.Start(); err != nil {
		log.Panicln(err)
	}
}

func mapConfigAuthItemToServerAuthItem(m map[string]ConfigAuthItem) (r map[string]server.AuthItem) {
	if m == nil {
		return r
	}
	r = make(map[string]server.AuthItem)
	for k, v := range m {
		r[k] = server.AuthItem{
			Modules: v.Modules,
			Allow:   v.Allow,
		}
	}
	return r
}
