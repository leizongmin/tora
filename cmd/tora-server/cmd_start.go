package main

import (
	"flag"
	"github.com/leizongmin/tora/server"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
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
		log.Fatalf("Load config failed: %s", err)
	}

	// 设置日志记录器
	if level, err := logrus.ParseLevel(c.Log.Level); err != nil {
		log.Errorf("Invalid log level: %s", c.Log.Level)
	} else {
		log.SetLevel(level)
	}
	switch c.Log.Format {
	case "text":
		log.Formatter = &logrus.TextFormatter{}
	case "json":
		log.Formatter = &logrus.JSONFormatter{}
	default:
		log.Errorf("Invalid log format: %s", c.Log.Format)
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
		log.Panicf("Try to start server failed: %s", err)
	}

	// 接收系统信号
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		sig := <-sigs
		log.Warnf("RECEIVED SIGNAL: %s", sig)
		s.Close()
		os.Exit(1)
	}()

	// 监听端口
	if err := s.Start(); err != nil {
		log.Error(err)
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
