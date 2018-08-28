package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"os"
)

func cmdInit(args []string) {
	var configFile string
	log := logrus.New()

	cmd := flag.NewFlagSet(CmdName, flag.ExitOnError)
	cmd.Usage = func() {
		printUsage(cmd)
	}
	cmd.StringVar(&configFile, "c", DefaultConfigFilePath, "set config file path")
	cmd.Parse(args)

	_, err := LoadConfigFile(configFile)
	if err == nil {
		log.Fatalf("Config file %s is already exists, please delete it firstly!", configFile)
	} else if os.IsNotExist(err) {
		_, err = CreateExampleConfigFile(configFile)
		if err != nil {
			log.Fatalf("Create config file failed: %s", err)
		} else {
			log.Warnf("Config file %s has been created", configFile)
		}
	} else {
		log.Fatalln(err)
	}
}
