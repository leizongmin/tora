package main

import (
	"flag"
	"fmt"
	"github.com/leizongmin/tora/server"
	"github.com/sirupsen/logrus"
	"os"
)

const DefaultConfigFilePath = "/etc/tora.yaml"

func main() {

	var configFile string
	var init bool
	log := logrus.New()

	// 解析命令行参数
	cmd := flag.NewFlagSet("tora-server", flag.ExitOnError)
	cmd.StringVar(&configFile, "c", DefaultConfigFilePath, "set c file path")
	cmd.BoolVar(&init, "init", false, "generate example config file")
	cmd.Usage = func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("tora/%s\n", server.Version))
		fmt.Fprintf(os.Stderr, "Usage: tora-server [-c filename]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		cmd.PrintDefaults()
	}
	cmd.Parse(os.Args[1:])

	// 读取配置文件
	c, err := LoadConfigFile(configFile)
	if err != nil {
		if os.IsNotExist(err) && init {
			c, err = CreateExampleConfigFile(configFile)
			if err != nil {
				log.Fatalf("Create config file failed: %s", err)
			} else {
				log.Warnf("Config file %s has been created", configFile)
			}
		} else {
			log.Fatalf("Load config failed: %s", err)
		}
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
