package main

import "flag"

func cmdDelete(args []string, cmd *flag.FlagSet, options baseOptions) {
	cmd.Parse(args)
}
