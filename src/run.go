package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

func runCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor run <instance-label> <executable> <arg-or-option>...")
		fatal(msg)
	}
	MustLoadSiteConfig()
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	site := getSite(ctx, api)
	instance, err := api.Instances.Get(site, ctx.Args()[0])
	if err != nil {
		fatal(err.Error())
	}
	if len(ctx.Args().Tail()) == 0 {
		usage("too few arguments")
	}
	endpoint, err := instance.Run("normal", ctx.Args().Tail())
	if err != nil {
		fatal(err.Error())
	}
	re := remoteExec{
		endpoint:   endpoint,
		enableTty:  true,
		httpClient: getHttpClient(ctx),
		tlsConfig:  getTLSConfig(ctx),
	}
	exitCode, err := re.execute()
	if err != nil {
		fatal(err.Error())
	}
	os.Exit(exitCode)
}
