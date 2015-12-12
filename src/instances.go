package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/olekukonko/tablewriter"
)

func instancesCreateCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor instances create [--kind] <label>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	site := getSite(ctx, api)
	kind := ctx.String("kind")
	instance := gondor.Instance{
		Site:  site.URL,
		Label: &ctx.Args()[0],
		Kind:  &kind,
	}
	if err := api.Instances.Create(&instance); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s instance has been created.", *instance.Label))
}

func instancesListCmd(ctx *cli.Context) {
	api := getAPIClient(ctx)
	site := getSite(ctx, api)
	instances, err := api.Instances.List(site.URL)
	if err != nil {
		fatal(err.Error())
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Label", "Kind"})
	for i := range instances {
		instance := instances[i]
		table.Append([]string{
			*instance.Label,
			*instance.Kind,
		})
	}
	table.Render()
}

func instancesDeleteCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor instances delete <label>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	site := getSite(ctx, api)
	label := ctx.Args()[0]
	instance, err := api.Instances.Get(*site.URL, label)
	if err != nil {
		fatal(err.Error())
	}
	if err := api.Instances.Delete(*instance.URL); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s instance has been deleted.", label))
}

func instancesEnvCmd(ctx *cli.Context) {
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	var err error
	var createMode bool
	var displayEnvVars, desiredEnvVars []*gondor.EnvironmentVariable
	if len(ctx.Args()) >= 1 {
		createMode = true
		for i := range ctx.Args() {
			arg := ctx.Args()[i]
			if strings.Contains(arg, "=") {
				parts := strings.Split(arg, "=")
				envVar := gondor.EnvironmentVariable{
					Instance: instance.URL,
					Key:      &parts[0],
					Value:    &parts[1],
				}
				desiredEnvVars = append(desiredEnvVars, &envVar)
			}
		}
	}
	if !createMode {
		displayEnvVars, err = api.EnvVars.ListByInstance(*instance.URL)
		if err != nil {
			fatal(err.Error())
		}
		for i := range displayEnvVars {
			envVar := displayEnvVars[i]
			fmt.Printf("%s=%s\n", *envVar.Key, *envVar.Value)
		}
	} else {
		if err := api.EnvVars.Create(desiredEnvVars); err != nil {
			fatal(err.Error())
		}
		for i := range desiredEnvVars {
			fmt.Printf("%s=%s\n", *desiredEnvVars[i].Key, *desiredEnvVars[i].Value)
		}
	}
}
