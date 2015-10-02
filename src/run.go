package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

func runCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor run [--instance] -- <executable> <arg-or-option>...")
		fatal(msg)
	}
	MustLoadSiteConfig()
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	endpoint, err := instance.Run("normal", ctx.Args())
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
