package main

import (
	"github.com/eldarion-gondor/cli"
)

var version string

func main() {
	cli := gondorcli.CLI{
		Name:     "gondor",
		LongName: "Gondor cloud",
		Version:  version,
		Author:   "Eldarion, Inc.",
		Email:    "development@eldarion.com",
	}
	cli.Prepare()
	cli.Run()
}
