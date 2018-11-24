package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func cmdInstall(args []string) {
	var configFile string
	var installType string
	var username string
	log := logrus.New()

	cmd := flag.NewFlagSet(CmdName, flag.ExitOnError)
	cmd.Usage = func() {
		printUsage(cmd)
	}
	cmd.StringVar(&configFile, "c", DefaultConfigFilePath, "set config file path")
	userInfo, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}
	username = userInfo.Username
	cmd.StringVar(&installType, "t", DefaultInstallType, "install type, you can choose: systemd")
	cmd.StringVar(&username, "u", userInfo.Username, "service run as specified user")
	cmd.Parse(args)

	// 读取配置文件
	c, err := LoadConfigFile(configFile)
	if err != nil {
		log.Fatalf("Load config failed: %s", err)
	}

	installService(log, configFile, &c, installType, username)
}

func installService(log *logrus.Logger, configFile string, config *Config, installType string, username string) {
	if installType != "systemd" {
		log.Fatalf("Unsupported service install type: %s", installType)
	}
	execPath, err := getSelfExecutable()
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
Type=notify
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

func getSelfExecutable() (string, error) {
	bin, err := os.Executable()
	if err != nil {
		return bin, err
	}
	return filepath.EvalSymlinks(bin)
}
