package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"os"
)

func cmdUninstall(args []string) {
	var configFile string
	var installType string
	log := logrus.New()

	cmd := flag.NewFlagSet(CmdName, flag.ExitOnError)
	cmd.Usage = func() {
		printUsage(cmd)
	}
	cmd.StringVar(&configFile, "c", DefaultConfigFilePath, "set config file path")
	cmd.StringVar(&installType, "t", DefaultInstallType, "install type, you can choose: systemd")
	cmd.Parse(args)

	// 读取配置文件
	c, err := LoadConfigFile(configFile)
	if err != nil {
		log.Fatalf("Load config failed: %s", err)
	}

	uninstallService(log, configFile, &c, installType)
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
