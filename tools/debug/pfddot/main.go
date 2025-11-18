package main

import (
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools/debug/pfddot/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
