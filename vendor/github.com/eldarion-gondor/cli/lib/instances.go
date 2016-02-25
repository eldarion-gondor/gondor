package gondorcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go/lib"
	"github.com/olekukonko/tablewriter"
)

func instancesCreateCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s instances create [--kind] <label>\n", c.Name)
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := c.GetAPIClient(ctx)
	site := c.GetSite(ctx)
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

func instancesListCmd(c *CLI, ctx *cli.Context) {
	api := c.GetAPIClient(ctx)
	site := c.GetSite(ctx)
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

func instancesDeleteCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s instances delete <label>\n", c.Name)
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := c.GetAPIClient(ctx)
	site := c.GetSite(ctx)
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

func instancesEnvCmd(c *CLI, ctx *cli.Context) {
	api := c.GetAPIClient(ctx)
	instance := c.GetInstance(ctx, nil)
	var err error
	var createMode bool
	var displayEnvVars, desiredEnvVars []*gondor.EnvironmentVariable
	if len(ctx.Args()) >= 1 {
		createMode = true
		for i := range ctx.Args() {
			arg := ctx.Args()[i]
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
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
