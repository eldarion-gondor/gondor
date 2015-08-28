package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/cli"
	"github.com/olekukonko/tablewriter"
)

func sitesListCmd(ctx *cli.Context) {
	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)

	var resourceGroup *gondor.ResourceGroup
	var err error
	if ctx.GlobalString("resource-group") != "" {
		resourceGroup, err = api.ResourceGroups.GetByName(ctx.GlobalString("resource-group"))
		if err != nil {
			fatal(err.Error())
		}
	}

	sites, err := api.Sites.List(resourceGroup)
	if err != nil {
		fatal(err.Error())
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Identifier"})
	for i := range sites {
		site := sites[i]
		table.Append([]string{
			fmt.Sprintf("%s/%s", site.ResourceGroup.Name, site.Name),
		})
	}
	table.Render()
}

func sitesCreateCmd(ctx *cli.Context) {
	var name string
	if len(ctx.Args()) == 1 {
		name = ctx.Args()[0]
	}
	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	resourceGroup := getResourceGroup(ctx, api)
	site := gondor.Site{
		ResourceGroup: resourceGroup,
		Name:          name,
	}
	if err := api.Sites.Create(&site); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%q site created.", fmt.Sprintf("%s/%s", site.ResourceGroup.Name, site.Name)))
}

func sitesDeleteCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor sites delete <site-name>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	resourceGroup := getResourceGroup(ctx, api)
	var site *gondor.Site
	site, err := api.Sites.Get(ctx.Args()[0], resourceGroup)
	if err != nil {
		fatal(err.Error())
	}
	if err := api.Sites.Delete(site); err != nil {
		fatal(err.Error())
	}
}
