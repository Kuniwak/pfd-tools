package main

import (
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools/pfddupmark/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
