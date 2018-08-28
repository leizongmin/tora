package main

import "flag"

func cmdGet(args []string, cmd *flag.FlagSet, options baseOptions) {
	cmd.Parse(args)
}
