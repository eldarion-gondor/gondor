package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/olekukonko/tablewriter"
)

func hostsListCmd(ctx *cli.Context) {
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
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
	usage := func(msg string) {
		fmt.Println("Usage: gondor hosts create [--instance] <hostname>")
		fatal(msg)
	}
	if len(ctx.Args()) < 2 {
		usage("too few arguments")
	}
	newHostName := ctx.Args()[0]
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	hostName := gondor.HostName{
		Instance: instance,
		Host:     newHostName,
	}
	if err := api.HostNames.Create(&hostName); err != nil {
		fatal(err.Error())
	}
}

func hostsDeleteCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor hosts delete [--instance] <hostname>")
		fatal(msg)
	}
	if len(ctx.Args()) < 2 {
		usage("too few arguments")
	}
	newHostName := ctx.Args()[0]
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	hostName := gondor.HostName{
		Instance: instance,
		Host:     newHostName,
	}
	if err := api.HostNames.Delete(&hostName); err != nil {
		fatal(err.Error())
	}
}
