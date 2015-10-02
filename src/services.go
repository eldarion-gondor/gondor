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
	MustLoadSiteConfig()
	usage := func(msg string) {
		fmt.Println("Usage: gondor services create [--name,--version,--instance] <service-kind>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	service := gondor.Service{
		Instance: instance,
		Name:     ctx.String("name"),
		Kind:     ctx.Args()[0],
	}
	if ctx.String("version") != "" {
		service.Version = ctx.String("version")
	}
	if err := api.Services.Create(&service); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s service has been created.", service.Kind))
}

func servicesListCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
	api := getAPIClient(ctx)
	instance := getInstance(ctx, api, nil)
	table := tablewriter.NewWriter(os.Stdout)
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
	table.Render()
}

func servicesDeleteCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
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
	service, err := api.Services.Get(instance, name)
	if err != nil {
		fatal(err.Error())
	}
	if err := api.Services.Delete(service); err != nil {
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
	service, err := api.Services.Get(instance, name)
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
					Service: service,
					Key:     parts[0],
					Value:   parts[1],
				}
				desiredEnvVars = append(desiredEnvVars, &envVar)
			}
		}
	}
	if !createMode {
		displayEnvVars, err = api.EnvVars.ListByService(service)
		for i := range displayEnvVars {
			envVar := displayEnvVars[i]
			fmt.Printf("%s=%s\n", envVar.Key, envVar.Value)
		}
	} else {
		if err := api.EnvVars.Create(desiredEnvVars); err != nil {
			fatal(err.Error())
		}
		for i := range desiredEnvVars {
			fmt.Printf("%s=%s\n", desiredEnvVars[i].Key, desiredEnvVars[i].Value)
		}
	}
}

func servicesScaleCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
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
	service, err := api.Services.Get(instance, name)
	if err != nil {
		fatal(err.Error())
	}
	if err := service.SetReplicas(replicas); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s service has been scaled to %d replicas.", name, replicas))
}

func servicesRestartCmd(ctx *cli.Context) {
	MustLoadSiteConfig()
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
	service, err := api.Services.Get(instance, name)
	if err != nil {
		fatal(err.Error())
	}
	if err := service.Restart(); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s service has been restarted.", name))
}