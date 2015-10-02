package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/skratchdot/open-golang/open"
)

func openCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	open.Run(fmt.Sprintf("https://%s/", instance.WebURL))
}
