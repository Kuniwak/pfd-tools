package main

import (
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools/pfdhelp/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
