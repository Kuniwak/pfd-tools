package main

import (
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools/pfdtable/cmd"
)

func main() {
	cmd.MainCommandByArgs(os.Args[1:], cli.NewProcInout())
}
