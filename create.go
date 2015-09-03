package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
)

func createCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
	usage := func(msg string) {
		fmt.Println("Usage: gondor create [--instance-kind,--service-name,--service-version] (<instance-label> | <service-kind>)")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	var instanceLabel string
	if len(ctx.Args()) > 1 {
		instanceLabel = ctx.Args()[0]
	}
	api := getAPIClient(ctx)
	site := getSite(ctx, api)
	if instanceLabel == "" {
		instance := gondor.Instance{
			Site:  site,
			Label: ctx.Args()[0],
			Kind:  ctx.String("instance-kind"),
		}
		if err := api.Instances.Create(&instance); err != nil {
			fatal(err.Error())
		}
		success(fmt.Sprintf("%s instance has been created.", instance.Label))
	} else {
		if len(ctx.Args()) < 2 {
			usage("too few arguments")
		}
		instance, err := api.Instances.Get(site, instanceLabel)
		if err != nil {
			fatal(err.Error())
		}
		service := gondor.Service{
			Instance: instance,
			Name:     ctx.String("service-name"),
			Kind:     ctx.Args()[1],
		}
		if ctx.String("version") != "" {
			service.Version = ctx.String("version")
		}
		if err := api.Services.Create(&service); err != nil {
			fatal(err.Error())
		}
		success(fmt.Sprintf("%s service has been created.", service.Kind))
	}
}
