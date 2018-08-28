package main

import (
	"fmt"
	"github.com/leizongmin/tora/server"
)

func cmdVersion(args []string) {
	fmt.Printf("tora-server/%s\n", server.Version)
}
