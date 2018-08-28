package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func cmdGet(args []string, cmd *flag.FlagSet, options baseOptions) {
	cmd.Parse(args)

	remotePath := cmd.Arg(0)
	if len(remotePath) < 1 {
		fmt.Println("Missing first argument <remotePath>")
	}
	remotePath = formatRemotePath(remotePath)
	fmt.Println("Remote Path:", remotePath)

	localPath := cmd.Arg(1)
	if len(localPath) > 0 {
		localPath = formatLocalPath(localPath)
	}
	fmt.Println("Local Path: ", localPath)

	client := NewClient(options.server, options.token)

	req, err := client.Get("file", remotePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	res, err := client.Response(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fileType := res.Header.Get("x-file-type")
	if fileType == "file" {
		if len(localPath) > 0 {
			writeToFile(localPath, res.Body)
		} else {
			fmt.Println()
			fmt.Println()
			writeToScreen(res.Body)
		}
	} else {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println()
		fmt.Println()
		fmt.Println(jsonPretty(body))
	}
}

func writeToFile(file string, data io.Reader) {
	fd, err := os.Create(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = io.Copy(fd, data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Write to:", file)
}

func writeToScreen(data io.Reader) {
	_, err := io.Copy(os.Stdout, data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println()
}
