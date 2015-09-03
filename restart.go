package main

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
)

func restartCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor restart <service-identifier>")
		fatal(msg)
	}
	MustLoadSiteConfig()
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	var instanceLabel, serviceName string
	arg := ctx.Args()[0]
	if strings.Contains(arg, "/") {
		parts := strings.Split(arg, "/")
		instanceLabel = parts[0]
		serviceName = parts[1]
	} else {
		usage(fmt.Sprintf("%q is not a service identifier", ctx.Args()[0]))
	}
	api := getAPIClient(ctx)
	site := getSite(ctx, api)
	instance, err := api.Instances.Get(site, instanceLabel)
	if err != nil {
		fatal(err.Error())
	}
	service, err := api.Services.Get(instance, serviceName)
	if err != nil {
		fatal(err.Error())
	}
	if err := service.Restart(); err != nil {
		fatal(err.Error())
	}
}
