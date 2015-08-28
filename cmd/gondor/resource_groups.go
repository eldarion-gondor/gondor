package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/cli"
	"github.com/olekukonko/tablewriter"
)

func resourceGroupListCmd(ctx *cli.Context) {
	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	resourceGroups, err := api.ResourceGroups.List()
	if err != nil {
		fatal(err.Error())
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Identifer"})
	for i := range resourceGroups {
		resourceGroup := resourceGroups[i]
		table.Append([]string{
			resourceGroup.Name,
		})
	}
	table.Render()
}
