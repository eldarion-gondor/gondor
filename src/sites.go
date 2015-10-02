package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/olekukonko/tablewriter"
)

func sitesListCmd(ctx *cli.Context) {
	api := getAPIClient(ctx)

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

func sitesInitCmd(ctx *cli.Context) {
	if _, err := os.Stat(configFilename); !os.IsNotExist(err) {
		fatal("site already initialized")
	}
	api := getAPIClient(ctx)
	resourceGroup := getResourceGroup(ctx, api)
	site := gondor.Site{
		ResourceGroup: resourceGroup,
		Name:          ctx.String("name"),
	}
	if err := api.Sites.Create(&site); err != nil {
		fatal(err.Error())
	}
	instance := gondor.Instance{
		Site:  &site,
		Label: "primary",
		Kind:  "production",
	}
	if err := api.Instances.Create(&instance); err != nil {
		fatal(err.Error())
	}
	sc := SiteConfig{
		Identifier: fmt.Sprintf("%s/%s", site.ResourceGroup.Name, site.Name),
	}
	buf, err := yaml.Marshal(sc)
	if err != nil {
		panic(err.Error())
	}
	if err := ioutil.WriteFile(configFilename, buf, 0644); err != nil {
		fatal(fmt.Sprintf("writing %s: %s", configFilename, err))
	}
	fmt.Printf("Wrote %s to your current directory.\nYour site is ready to be deployed. To deploy, run:\n\n\tg3a deploy primary master\n\nDon't forget to commit %s before deploying.\n", configFilename, configFilename)
}

func sitesCreateCmd(ctx *cli.Context) {
	var name string
	if len(ctx.Args()) == 1 {
		name = ctx.Args()[0]
	}
	api := getAPIClient(ctx)
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
	api := getAPIClient(ctx)
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
