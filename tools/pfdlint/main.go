package main

import (
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools/pfdlint/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
