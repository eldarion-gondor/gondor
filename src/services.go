package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/olekukonko/tablewriter"
)

func servicesCreateCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor services create [--name,--version,--instance] <service-kind>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	name := ctx.String("name")
	service := gondor.Service{
		Instance: instance.URL,
		Name:     &name,
		Kind:     &ctx.Args()[0],
	}
	if ctx.String("version") != "" {
		version := ctx.String("version")
		service.Version = &version
	}
	if err := api.Services.Create(&service); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s service has been created.", *service.Kind))
}

func servicesListCmd(ctx *cli.Context) {
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	services, err := api.Services.List(&*instance.URL)
	if err != nil {
		fatal(err.Error())
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Kind", "Replicas", "Web URL", "State"})
	for i := range services {
		service := services[i]
		var webURL string
		if service.WebURL != nil {
			webURL = *service.WebURL
		}
		table.Append([]string{
			*service.Name,
			*service.Kind,
			strconv.Itoa(*service.Replicas),
			webURL,
			*service.State,
		})
	}
	table.Render()
}

func servicesDeleteCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor services delete <name>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	name := ctx.Args()[0]
	service, err := api.Services.Get(*instance.URL, name)
	if err != nil {
		fatal(err.Error())
	}
	if err := api.Services.Delete(*service.URL); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s service has been deleted.", name))
}

func servicesEnvCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor services env <name>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	name := ctx.Args()[0]
	service, err := api.Services.Get(*instance.URL, name)
	if err != nil {
		fatal(err.Error())
	}
	var createMode bool
	var displayEnvVars, desiredEnvVars []*gondor.EnvironmentVariable
	if len(ctx.Args()) >= 2 {
		createMode = true
		for i := range ctx.Args() {
			arg := ctx.Args()[i]
			if strings.Contains(arg, "=") {
				parts := strings.Split(arg, "=")
				envVar := gondor.EnvironmentVariable{
					Service: service.URL,
					Key:     &parts[0],
					Value:   &parts[1],
				}
				desiredEnvVars = append(desiredEnvVars, &envVar)
			}
		}
	}
	if !createMode {
		displayEnvVars, err = api.EnvVars.ListByService(*service.URL)
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

func servicesScaleCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor services scale --replicas=N <name>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	name := ctx.Args()[0]
	replicas := ctx.Int("replicas")
	service, err := api.Services.Get(*instance.URL, name)
	if err != nil {
		fatal(err.Error())
	}
	if err := service.SetReplicas(replicas); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s service has been scaled to %d replicas.", name, replicas))
}

func servicesRestartCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor services restart <name>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	name := ctx.Args()[0]
	service, err := api.Services.Get(*instance.URL, name)
	if err != nil {
		fatal(err.Error())
	}
	if err := service.Restart(); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s service has been restarted.", name))
}
