package main

import (
	"os"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/olekukonko/tablewriter"
)

func listCmd(ctx *cli.Context) {
	var instanceLabel string
	if len(ctx.Args()) == 1 {
		instanceLabel = ctx.Args()[0]
	}
	api := getAPIClient(ctx)
	site := getSite(ctx, api)
	table := tablewriter.NewWriter(os.Stdout)
	if instanceLabel == "" {
		// show instances of a site
		table.SetHeader([]string{"Label", "Kind", "URL"})
		for i := range site.Instances {
			instance := site.Instances[i]
			table.Append([]string{
				instance.Label,
				instance.Kind,
				instance.WebURL,
			})
		}
	} else {
		// show services of an instance
		instance, err := api.Instances.Get(site, instanceLabel)
		if err != nil {
			fatal(err.Error())
		}
		table.SetHeader([]string{"Name", "Kind", "Replicas", "State"})
		for i := range instance.Services {
			service := instance.Services[i]
			table.Append([]string{
				service.Name,
				service.Kind,
				strconv.Itoa(service.Replicas),
				service.State,
			})
		}
	}
	table.Render()
}
