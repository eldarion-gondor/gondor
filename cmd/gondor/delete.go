package main

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/cli"
)

func deleteCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
	usage := func(msg string) {
		fmt.Println("Usage: gondor delete (<instance-label> | <service-identifier>)")
		fatal(msg)
	}
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
		instanceLabel = arg
	}
	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	site := getSite(ctx, api)
	if instanceLabel != "" && serviceName == "" {
		instance, err := api.Instances.Get(site, instanceLabel)
		if err != nil {
			fatal(err.Error())
		}
		if err := api.Instances.Delete(instance); err != nil {
			fatal(err.Error())
		}
	} else {
		instance, err := api.Instances.Get(site, instanceLabel)
		if err != nil {
			fatal(err.Error())
		}
		service, err := api.Services.Get(instance, serviceName)
		if err != nil {
			fatal(err.Error())
		}
		if err := api.Services.Delete(service); err != nil {
			fatal(err.Error())
		}
	}
}
