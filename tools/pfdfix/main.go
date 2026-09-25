package main

import (
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools/pfdfix/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
