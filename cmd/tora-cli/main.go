package main

import (
	"flag"
	"fmt"
	"github.com/leizongmin/tora/server"
	"os"
	"runtime"
)

const CmdName = "tora-cli"

func main() {
	// 获取子命令
	var cmdType string
	var args []string
	if len(os.Args) < 2 {
		args = os.Args[1:]
		cmdType = "help"
	} else {
		cmdType = os.Args[1]
		if cmdType[0:1] == "-" {
			cmdType = "help"
			args = os.Args[1:]
		} else {
			args = os.Args[2:]
		}
	}

	cmd := flag.NewFlagSet(CmdName, flag.ExitOnError)
	cmd.Usage = func() {
		printUsage(cmd)
	}
	var options baseOptions
	cmd.StringVar(&options.server, "s", "http://127.0.0.1"+server.DefaultListenAddr, "Remote server address")
	cmd.StringVar(&options.token, "t", "", "Auth token")

	switch cmdType {
	case "put":
		cmdPut(args, cmd, options)
	case "delete":
		cmdDelete(args, cmd, options)
	case "get":
		cmdGet(args, cmd, options)
	case "help":
		printUsage(cmd)
	default:
		printUsage(cmd)
		os.Exit(1)
	}
}

func printUsage(cmd *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s/%s for %s\n\n", CmdName, server.Version, runtime.GOOS)
	fmt.Fprintf(os.Stderr, "Usage: \n")
	fmt.Fprintf(os.Stderr, "    %s [-s server] [-t token]\n", CmdName)
	fmt.Fprintf(os.Stderr, "        put <remotePath> <localPath>      Put file or directory to remote server\n")
	fmt.Fprintf(os.Stderr, "        delete <remotePath>               Delete file or directory from remote server\n")
	fmt.Fprintf(os.Stderr, "        get <remotePath>                  Get file from remote server\n")
	if cmd != nil {
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		cmd.PrintDefaults()
	}
}

type baseOptions struct {
	server string
	token  string
}
