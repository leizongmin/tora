package main

import (
	"flag"
	"fmt"
)

func cmdGet(args []string, cmd *flag.FlagSet, options baseOptions) {
	cmd.Parse(args)
	fmt.Println(cmd.Arg(0))
	fmt.Println(args)
}
