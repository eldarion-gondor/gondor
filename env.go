package main

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
)

func envCmd(ctx *cli.Context) {
	MustLoadSiteConfig()

	var err error
	var createMode bool
	var site *gondor.Site
	var instance *gondor.Instance
	var service *gondor.Service
	var displayEnvVars, desiredEnvVars []*gondor.EnvironmentVariable
	var scope, instanceLabel, serviceName string

	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	site = getSite(ctx, api)

	if len(ctx.Args()) >= 1 {
		arg := ctx.Args()[0]
		if strings.Contains(arg, "/") {
			parts := strings.Split(arg, "/")
			instanceLabel = parts[0]
			serviceName = parts[1]
		} else {
			instanceLabel = arg
		}
		// load up any desired environmental variables
		if len(ctx.Args()) > 1 {
			createMode = true
			for i := range ctx.Args() {
				arg := ctx.Args()[i]
				if strings.Contains(arg, "=") {
					parts := strings.Split(arg, "=")
					envVar := gondor.EnvironmentVariable{
						Key:   parts[0],
						Value: parts[1],
					}
					desiredEnvVars = append(desiredEnvVars, &envVar)
				}
			}
		}
	}

	// deterimine scope and fetch objects from API
	if instanceLabel == "" {
		scope = "site"
	} else {
		instance, err = api.Instances.Get(site, instanceLabel)
		if err != nil {
			fatal(err.Error())
		}
		if serviceName != "" {
			service, err = api.Services.Get(instance, serviceName)
			if err != nil {
				fatal(err.Error())
			}
			scope = "service"
		} else {
			scope = "instance"
		}
	}

	if !createMode {
		switch scope {
		case "site":
			displayEnvVars, err = api.EnvVars.ListBySite(site)
			break
		case "instance":
			displayEnvVars, err = api.EnvVars.ListByInstance(instance)
		case "service":
			displayEnvVars, err = api.EnvVars.ListByService(service)
		}
		if err != nil {
			fatal(err.Error())
		}
		for i := range displayEnvVars {
			envVar := displayEnvVars[i]
			fmt.Printf("%s=%s\n", envVar.Key, envVar.Value)
		}
	} else {
		switch scope {
		case "site":
			for i := range desiredEnvVars {
				desiredEnvVars[i].Site = site
			}
			break
		case "instance":
			for i := range desiredEnvVars {
				desiredEnvVars[i].Instance = instance
			}
			break
		case "service":
			for i := range desiredEnvVars {
				desiredEnvVars[i].Service = service
			}
			break
		}
		if err := api.EnvVars.Create(desiredEnvVars); err != nil {
			fatal(err.Error())
		}
		for i := range desiredEnvVars {
			fmt.Printf("%s=%s\n", desiredEnvVars[i].Key, desiredEnvVars[i].Value)
		}
	}
}
