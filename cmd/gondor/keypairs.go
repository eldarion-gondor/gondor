package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/cli"
	"github.com/olekukonko/tablewriter"
)

func keypairsListCmd(ctx *cli.Context) {
	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	resourceGroup := getResourceGroup(ctx, api)

	keypairs, err := api.KeyPairs.List(resourceGroup)
	if err != nil {
		fatal(err.Error())
	}

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{}
	if resourceGroup == nil {
		header = append(header, "Resource Group")
	}
	header = append(header, "Name")
	table.SetHeader(header)
	for i := range keypairs {
		keypair := keypairs[i]
		row := []string{}
		if resourceGroup == nil {
			row = append(row, keypair.ResourceGroup.Name)
		}
		row = append(row, keypair.Name)
		table.Append(row)
	}
	table.Render()
}

func keypairsCreateCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor keypairs create --name=<keypair-name> <private-key-path> <certificate-path>")
		fatal(msg)
	}

	if len(ctx.Args()) < 2 {
		usage("too few arguments")
	}
	if !ctx.IsSet("name") || ctx.String("name") == "" {
		usage("--name is required")
	}

	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	resourceGroup := getResourceGroup(ctx, api)

	privateKeyPath := ctx.Args()[0]
	certPath := ctx.Args()[1]

	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		fatal(err.Error())
	}
	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		fatal(err.Error())
	}

	keypair := gondor.KeyPair{
		ResourceGroup: resourceGroup,
		Name:          ctx.String("name"),
		Key:           privateKey,
		Certificate:   cert,
	}
	if err := api.KeyPairs.Create(&keypair); err != nil {
		fatal(err.Error())
	}
	success("keypair created.")
}

func keypairsAttachCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor keypairs attach --keypair=<keypair-name> --service=<service-identifer>")
		fatal(msg)
	}

	if ctx.String("keypair") == "" {
		usage("--keypair is required")
	}
	if ctx.String("service") == "" {
		usage("--service is required")
	}

	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	resourceGroup := getResourceGroup(ctx, api)
	site := getSite(ctx, api)

	keypair, err := api.KeyPairs.GetByName(ctx.String("keypair"), resourceGroup)
	if err != nil {
		fatal(err.Error())
	}

	var instanceLabel, serviceName string

	if strings.Contains(ctx.String("service"), "/") {
		parts := strings.Split(ctx.String("service"), "/")
		instanceLabel = parts[0]
		serviceName = parts[1]
	} else {
		fatal("invalid --service value")
	}
	instance, err := api.Instances.Get(site, instanceLabel)
	if err != nil {
		fatal(err.Error())
	}
	service, err := api.Services.Get(instance, serviceName)
	if err != nil {
		fatal(err.Error())
	}

	// attach the keypair to the service
	service = &gondor.Service{
		URL:     service.URL,
		KeyPair: keypair,
	}
	if err := api.Services.Update(*service); err != nil {
		fatal(err.Error())
	}
	success("keypair attached.")
}

func keypairsDetachCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor keypairs detach --service=<service-identifer>")
		fatal(msg)
	}

	if ctx.String("service") == "" {
		usage("--service is required")
	}

	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	site := getSite(ctx, api)

	var instanceLabel, serviceName string

	if strings.Contains(ctx.String("service"), "/") {
		parts := strings.Split(ctx.String("service"), "/")
		instanceLabel = parts[0]
		serviceName = parts[1]
	} else {
		usage(fmt.Sprintf("%q is not a service identifier", ctx.String("service")))
	}
	instance, err := api.Instances.Get(site, instanceLabel)
	if err != nil {
		fatal(err.Error())
	}
	service, err := api.Services.Get(instance, serviceName)
	if err != nil {
		fatal(err.Error())
	}

	// detach the keypair from the service using custom struct to allow an empty
	// keypair
	if err := service.DetachKeyPair(); err != nil {
		fatal(err.Error())
	}
	success("keypair detached.")
}

func keypairsDeleteCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor keypairs delete <keypair-name>")
		fatal(msg)
	}

	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}

	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	resourceGroup := getResourceGroup(ctx, api)

	keypair, err := api.KeyPairs.GetByName(ctx.Args()[0], resourceGroup)
	if err != nil {
		fatal(err.Error())
	}

	if err := api.KeyPairs.Delete(keypair); err != nil {
		fatal(err.Error())
	}

	success("keypair deleted.")
}
