package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/skratchdot/open-golang/open"
)

func openCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor open <instance-label>")
		fatal(msg)
	}
	MustLoadSiteConfig()
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	instanceLabel := ctx.Args()[0]

	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	site := getSite(ctx, api)
	instance, err := api.Instances.Get(site, instanceLabel)
	if err != nil {
		fatal(err.Error())
	}

	open.Run(fmt.Sprintf("https://%s/", instance.WebURL))
}
