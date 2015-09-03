package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/olekukonko/tablewriter"
)

func hostsListCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
	usage := func(msg string) {
		fmt.Println("Usage: gondor hosts list <instance-label>")
		fatal(msg)
	}
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

	hostNames, err := api.HostNames.List(instance)
	if err != nil {
		fatal(err.Error())
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Host"})
	for i := range hostNames {
		hostName := hostNames[i]
		table.Append([]string{
			hostName.Host,
		})
	}
	table.Render()
}

func hostsCreateCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
	usage := func(msg string) {
		fmt.Println("Usage: gondor hosts create <instance-label> <hostname>")
		fatal(msg)
	}
	if len(ctx.Args()) < 2 {
		usage("too few arguments")
	}

	instanceLabel := ctx.Args()[0]
	newHostName := ctx.Args()[1]

	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	site := getSite(ctx, api)
	instance, err := api.Instances.Get(site, instanceLabel)
	if err != nil {
		fatal(err.Error())
	}

	hostName := gondor.HostName{
		Instance: instance,
		Host:     newHostName,
	}
	if err := api.HostNames.Create(&hostName); err != nil {
		fatal(err.Error())
	}
}

func hostsDeleteCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
	usage := func(msg string) {
		fmt.Println("Usage: gondor hosts delete <instance-label> <hostname>")
		fatal(msg)
	}
	if len(ctx.Args()) < 2 {
		usage("too few arguments")
	}

	instanceLabel := ctx.Args()[0]
	newHostName := ctx.Args()[1]

	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	site := getSite(ctx, api)

	// lookup instance
	instance, err := api.Instances.Get(site, instanceLabel)
	if err != nil {
		fatal(err.Error())
	}

	hostName := gondor.HostName{
		Instance: instance,
		Host:     newHostName,
	}
	if err := api.HostNames.Delete(&hostName); err != nil {
		fatal(err.Error())
	}
}
