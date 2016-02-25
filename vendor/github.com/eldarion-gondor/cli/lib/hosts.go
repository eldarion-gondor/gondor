package gondorcli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go/lib"
	"github.com/olekukonko/tablewriter"
)

func hostsListCmd(c *CLI, ctx *cli.Context) {
	api := c.GetAPIClient(ctx)
	instance := c.GetInstance(ctx, nil)
	hostNames, err := api.HostNames.List(instance.URL)
	if err != nil {
		fatal(err.Error())
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Host"})
	for i := range hostNames {
		hostName := hostNames[i]
		table.Append([]string{
			*hostName.Host,
		})
	}
	table.Render()
}

func hostsCreateCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s hosts create [--instance] <hostname>\n", c.Name)
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	newHostName := ctx.Args()[0]
	api := c.GetAPIClient(ctx)
	instance := c.GetInstance(ctx, nil)
	hostName := gondor.HostName{
		Instance: instance.URL,
		Host:     &newHostName,
	}
	if err := api.HostNames.Create(&hostName); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s host has been created.", *hostName.Host))
}

func hostsDeleteCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s hosts delete [--instance] <hostname>\n", c.Name)
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	newHostName := ctx.Args()[0]
	api := c.GetAPIClient(ctx)
	instance := c.GetInstance(ctx, nil)
	hostName := gondor.HostName{
		Instance: instance.URL,
		Host:     &newHostName,
	}
	if err := api.HostNames.Delete(&hostName); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s instance has been deleted.", *hostName.Host))
}
