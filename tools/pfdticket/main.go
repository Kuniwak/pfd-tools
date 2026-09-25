package main

import (
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools/pfdticket/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
