package gondorcli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go/lib"
	"github.com/olekukonko/tablewriter"
)

func keypairsListCmd(c *CLI, ctx *cli.Context) {
	api := c.GetAPIClient(ctx)
	resourceGroup := c.GetResourceGroup(ctx)
	keypairs, err := api.KeyPairs.List(resourceGroup.URL)
	if err != nil {
		fatal(err.Error())
	}
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{}
	header = append(header, "Name")
	table.SetHeader(header)
	for i := range keypairs {
		keypair := keypairs[i]
		row := []string{}
		row = append(row, *keypair.Name)
		table.Append(row)
	}
	table.Render()
}

func keypairsCreateCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s keypairs create --name=<keypair-name> <private-key-path> <certificate-path>\n", c.Name)
		fatal(msg)
	}
	if len(ctx.Args()) < 2 {
		usage("too few arguments")
	}
	if !ctx.IsSet("name") || ctx.String("name") == "" {
		usage("--name is required")
	}
	api := c.GetAPIClient(ctx)
	resourceGroup := c.GetResourceGroup(ctx)
	name := ctx.String("name")
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
		ResourceGroup: resourceGroup.URL,
		Name:          &name,
		Key:           privateKey,
		Certificate:   cert,
	}
	if err := api.KeyPairs.Create(&keypair); err != nil {
		fatal(err.Error())
	}
	success("keypair created.")
}

func keypairsAttachCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s keypairs attach [--instance] --keypair=<keypair-name> --service=<name>\n", c.Name)
		fatal(msg)
	}
	if ctx.String("keypair") == "" {
		usage("--keypair is required")
	}
	if ctx.String("service") == "" {
		usage("--service is required")
	}
	api := c.GetAPIClient(ctx)
	resourceGroup := c.GetResourceGroup(ctx)
	keypair, err := api.KeyPairs.GetByName(ctx.String("keypair"), resourceGroup.URL)
	if err != nil {
		fatal(err.Error())
	}
	instance := c.GetInstance(ctx, nil)
	service, err := api.Services.Get(*instance.URL, ctx.String("service"))
	if err != nil {
		fatal(err.Error())
	}
	// attach the keypair to the service
	service = &gondor.Service{
		URL:     service.URL,
		KeyPair: keypair.URL,
	}
	if err := api.Services.Update(*service); err != nil {
		fatal(err.Error())
	}
	success("keypair attached.")
}

func keypairsDetachCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s keypairs detach [--instance] --service=<name>\n", c.Name)
		fatal(msg)
	}
	if ctx.String("service") == "" {
		usage("--service is required")
	}
	api := c.GetAPIClient(ctx)
	instance := c.GetInstance(ctx, nil)
	service, err := api.Services.Get(*instance.URL, ctx.String("service"))
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

func keypairsDeleteCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s keypairs delete <keypair-name>\n", c.Name)
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := c.GetAPIClient(ctx)
	resourceGroup := c.GetResourceGroup(ctx)
	keypair, err := api.KeyPairs.GetByName(ctx.Args()[0], resourceGroup.URL)
	if err != nil {
		fatal(err.Error())
	}
	if err := api.KeyPairs.Delete(*keypair.URL); err != nil {
		fatal(err.Error())
	}
	success("keypair deleted.")
}
