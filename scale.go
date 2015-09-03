package main

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
)

func scaleCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
	usage := func(msg string) {
		fmt.Println("Usage: gondor scale --replicas=N <service-identifier>")
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
		usage("invalid service identifier value")
	}
	if !ctx.IsSet("replicas") {
		usage("--replicas is required")
	}
	replicas := ctx.Int("replicas")
	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	site := getSite(ctx, api)
	instance, err := api.Instances.Get(site, instanceLabel)
	if err != nil {
		fatal(err.Error())
	}
	service, err := api.Services.Get(instance, serviceName)
	if err != nil {
		fatal(err.Error())
	}
	if err := service.SetReplicas(replicas); err != nil {
		fatal(err.Error())
	}
}
