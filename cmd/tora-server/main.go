package main

import (
	"flag"
	"fmt"
	"github.com/leizongmin/tora/server"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

const DefaultConfigFilePath = "/etc/tora.yaml"
const DefaultInstallType = "systemd"
const SystemdServiceFilePath = "/lib/systemd/system/tora.service"

func main() {

	var printVersion bool
	var configFile string
	var init bool
	var install bool
	var uninstall bool
	var installType string
	var username string
	log := logrus.New()

	// 解析命令行参数
	cmd := flag.NewFlagSet("tora-server", flag.ExitOnError)
	cmd.BoolVar(&printVersion, "v", false, "print printVersion info")

	cmd.StringVar(&configFile, "c", DefaultConfigFilePath, "set c file path")
	cmd.BoolVar(&init, "init", false, "generate example config file")

	userInfo, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}
	username = userInfo.Username
	cmd.BoolVar(&install, "install", false, "install system service")
	cmd.BoolVar(&uninstall, "uninstall", false, "uninstall system service")
	cmd.StringVar(&installType, "t", DefaultInstallType, "install type, you can choose: systemd")
	cmd.StringVar(&username, "u", userInfo.Username, "service run as specified user")

	cmd.Usage = func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("tora/%server\n", server.Version))
		fmt.Fprintf(os.Stderr, "Usage: tora-server [-c filename] [-init]\n")
		fmt.Fprintf(os.Stderr, "       tora-server -install [-t systemd] [-c filename]\n")
		fmt.Fprintf(os.Stderr, "       tora-server -uninstall [-t systemd] [-c filename]\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		cmd.PrintDefaults()
	}
	cmd.Parse(os.Args[1:])

	if printVersion {
		fmt.Printf("tora/%server\n", server.Version)
		return
	}

	log.Infof("PID: %d", os.Getpid())

	// 读取配置文件
	c, err := LoadConfigFile(configFile)
	if err != nil {
		// 如果文件不存在且指定了 -init 选项则创建默认文件
		if os.IsNotExist(err) && init {
			c, err = CreateExampleConfigFile(configFile)
			if err != nil {
				log.Fatalf("Create config file failed: %server", err)
			} else {
				log.Warnf("Config file %server has been created", configFile)
			}
		} else {
			log.Fatalf("Load config failed: %server", err)
		}
	}

	// 设置日志记录器
	if level, err := logrus.ParseLevel(c.Log.Level); err != nil {
		log.Errorf("Invalid log level: %server", c.Log.Level)
	} else {
		log.SetLevel(level)
	}
	switch c.Log.Format {
	case "text":
		log.Formatter = &logrus.TextFormatter{}
	case "json":
		log.Formatter = &logrus.JSONFormatter{}
	default:
		log.Errorf("Invalid log format: %server", c.Log.Format)
	}

	// 安装为系统服务
	if install {
		installService(log, configFile, &c, installType, username)
		return
	}
	if uninstall {
		uninstallService(log, configFile, &c, installType)
		return
	}

	// 创建服务器实例
	server, err := server.NewServer(server.Options{
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
		log.Panicf("Try to start server failed: %server", err)
	}

	// 接收系统信号
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		s := <-sigs
		log.Warnf("RECEIVED SIGNAL: %s", s)
		server.Close()
		os.Exit(1)
	}()

	// 监听端口
	if err := server.Start(); err != nil {
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

func installService(log *logrus.Logger, configFile string, config *Config, installType string, username string) {
	if installType != "systemd" {
		log.Fatalf("Unsupported service install type: %s", installType)
	}
	execPath, err := getExecutable()
	if err != nil {
		log.Fatalln(err)
	}
	configFile, err = filepath.Abs(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	tpl := strings.TrimSpace(`
[Unit]
Description=tora-server

[Service]
Type=simple
ExecStart=%s -c %s
WatchdogSec=30s
Restart=on-failure
User=%s
Group=%s

[Install]
WantedBy=multi-user.target
	`)
	code := fmt.Sprintf(tpl, execPath, configFile, username, username)
	err = ioutil.WriteFile(SystemdServiceFilePath, []byte(code), 0644)
	if err != nil {
		log.Fatalln(err)
	}
	log.Infof("Created service file: %s", SystemdServiceFilePath)
}

func uninstallService(log *logrus.Logger, configFile string, config *Config, installType string) {
	if installType != "systemd" {
		log.Fatalf("Unsupported service install type: %s", installType)
	}
	err := os.Remove(SystemdServiceFilePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	log.Infof("Deleted service file: %s", SystemdServiceFilePath)
}

func getExecutable() (string, error) {
	bin, err := os.Executable()
	if err != nil {
		return bin, err
	}
	return filepath.EvalSymlinks(bin)
}
