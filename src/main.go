package main

import (
	"github.com/eldarion-gondor/cli"
)

var version string

func main() {
	cli := gondorcli.CLI{
		Name:    "gondor",
		Version: version,
		Author:  "Eldarion, Inc.",
		Email:   "development@eldarion.com",
		Usage:   "command-line tool for interacting with the Gondor API",
	}
	cli.Run()
}
