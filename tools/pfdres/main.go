package main

import (
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools/pfdres/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
