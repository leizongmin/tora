package main

import (
	"flag"
	"fmt"
	"github.com/leizongmin/tora/server"
	"os"
	"runtime"
)

const DefaultConfigFilePath = "/etc/tora.yaml"
const DefaultInstallType = "systemd"
const SystemdServiceFilePath = "/lib/systemd/system/tora.service"
const CmdName = "tora-server"

func main() {

	// 获取子命令
	var cmdType string
	var args []string
	if len(os.Args) < 2 {
		args = os.Args[1:]
		cmdType = "start"
	} else {
		cmdType = os.Args[1]
		if cmdType[0:1] == "-" {
			cmdType = "start"
			args = os.Args[1:]
		} else {
			args = os.Args[2:]
		}
	}

	switch cmdType {
	case "start":
		cmdStart(args)
	case "version":
		cmdVersion(args)
	case "install":
		cmdInstall(args)
	case "uninstall":
		cmdUninstall(args)
	case "init":
		cmdInit(args)
	case "help":
		printUsage(nil)
	default:
		printUsage(nil)
		os.Exit(1)
	}
}

func printUsage(cmd *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("tora-server/%s for %s\n\n", server.Version, runtime.GOOS))
	fmt.Fprintf(os.Stderr, "Usage: tora-server start [-c filename]                      Start service\n")
	fmt.Fprintf(os.Stderr, "       tora-server version                                  Print version info\n")
	fmt.Fprintf(os.Stderr, "       tora-server install [-t systemd] [-c filename]       Install service\n")
	fmt.Fprintf(os.Stderr, "       tora-server uninstall [-t systemd] [-c filename]     Uninstall service\n")
	fmt.Fprintf(os.Stderr, "       tora-server init [-c filename]                       Create example config file\n")
	fmt.Fprintf(os.Stderr, "       tora-server help                                     Print usage\n")
	if cmd != nil {
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		cmd.PrintDefaults()
	}
}
