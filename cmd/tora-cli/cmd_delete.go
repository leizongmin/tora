package main

import (
	"flag"
	"fmt"
	"os"
)

func cmdDelete(args []string, cmd *flag.FlagSet, options baseOptions) {
	cmd.Parse(args)

	remotePath := cmd.Arg(0)
	if len(remotePath) < 1 {
		fmt.Println("Missing first argument <remotePath>")
	}
	remotePath = formatRemotePath(remotePath)
	fmt.Println("Remote Path:", remotePath)

	client := NewClient(options.server, options.token)

	req, err := client.Delete("file", remotePath, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, data, err := client.ResponseBytes(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println()
	fmt.Println(string(data))
}
