package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

func runCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor run [--instance] <service-name> -- <executable> <arg-or-option>...")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	service, err := api.Services.Get(*instance.URL, ctx.Args()[0])
	if err != nil {
		fatal(err.Error())
	}
	endpoint, err := service.Run(ctx.Args()[1:])
	if err != nil {
		fatal(err.Error())
	}
	re := remoteExec{
		endpoint:      endpoint,
		enableTty:     true,
		httpClient:    getHttpClient(ctx),
		tlsConfig:     getTLSConfig(ctx),
		showAttaching: true,
	}
	exitCode, err := re.execute()
	if err != nil {
		fatal(err.Error())
	}
	os.Exit(exitCode)
}
