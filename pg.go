package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

func pgRunCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor pg run <instance-label> <executable> <arg-or-option>...")
		fatal(msg)
	}
	MustLoadSiteConfig()
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	instanceLabel := ctx.Args()[0]
	api := getAPIClient(ctx)
	site := getSite(ctx, api)
	instance, err := api.Instances.Get(site, instanceLabel)
	if err != nil {
		fatal(err.Error())
	}
	if len(ctx.Args().Tail()) == 0 {
		usage("too few arguments")
	}
	endpoint, err := instance.Run("pg", ctx.Args().Tail())
	if err != nil {
		fatal(err.Error())
	}
	exitCode, err := remoteExec(endpoint, true)
	if err != nil {
		fatal(err.Error())
	}
	os.Exit(exitCode)
}
