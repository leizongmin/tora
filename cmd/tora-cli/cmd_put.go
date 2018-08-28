package main

import "flag"

func cmdPut(args []string, cmd *flag.FlagSet, options baseOptions) {
	cmd.Parse(args)
}
