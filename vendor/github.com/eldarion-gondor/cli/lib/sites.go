package gondorcli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go/lib"
	"github.com/olekukonko/tablewriter"
)

func sitesListCmd(c *CLI, ctx *cli.Context) {
	api := c.GetAPIClient(ctx)
	resourceGroup := c.GetResourceGroup(ctx)

	sites, err := api.Sites.List(&*resourceGroup.URL)
	if err != nil {
		fatal(err.Error())
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name"})
	for i := range sites {
		site := sites[i]
		table.Append([]string{
			*site.Name,
		})
	}
	table.Render()
}

func sitesInitCmd(c *CLI, ctx *cli.Context) {
	if _, err := os.Stat(configFilename); !os.IsNotExist(err) {
		fatal("site already initialized")
	}
	api := c.GetAPIClient(ctx)
	resourceGroup := c.GetResourceGroup(ctx)
	name := ctx.String("name")
	site := gondor.Site{
		ResourceGroup: resourceGroup.URL,
		Name:          &name,
	}
	if err := api.Sites.Create(&site); err != nil {
		fatal(err.Error())
	}
	label := "primary"
	kind := "production"
	instance := gondor.Instance{
		Site:  site.URL,
		Label: &label,
		Kind:  &kind,
	}
	if err := api.Instances.Create(&instance); err != nil {
		fatal(err.Error())
	}
	serviceKind := "web"
	service := gondor.Service{
		Instance: instance.URL,
		Name:     &serviceKind,
		Kind:     &serviceKind,
	}
	if err := api.Services.Create(&service); err != nil {
		fatal(err.Error())
	}
	sc := SiteConfig{
		Identifier: fmt.Sprintf("%s/%s", *resourceGroup.Name, *site.Name),
		Branches:   map[string]string{"master": label},
		Deploy: &DeployConfig{
			Services: []string{*service.Name},
		},
	}
	buf, err := yaml.Marshal(sc)
	if err != nil {
		panic(err.Error())
	}
	if err := ioutil.WriteFile(configFilename, buf, 0644); err != nil {
		fatal(fmt.Sprintf("writing %s: %s", configFilename, err))
	}
	fmt.Printf("Wrote %s to your current directory.\nYour site is ready to be deployed. To deploy, run:\n\n\t%s deploy\n\nDon't forget to commit %s before deploying.\n", configFilename, c.Name, configFilename)
}

func sitesCreateCmd(c *CLI, ctx *cli.Context) {
	var name string
	if len(ctx.Args()) == 1 {
		name = ctx.Args()[0]
	}
	api := c.GetAPIClient(ctx)
	resourceGroup := c.GetResourceGroup(ctx)
	site := gondor.Site{
		ResourceGroup: resourceGroup.URL,
		Name:          &name,
	}
	if err := api.Sites.Create(&site); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%q site created.", fmt.Sprintf("%s/%s", *resourceGroup.Name, *site.Name)))
}

func sitesDeleteCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s sites delete <site-name>\n", c.Name)
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	api := c.GetAPIClient(ctx)
	resourceGroup := c.GetResourceGroup(ctx)
	var site *gondor.Site
	site, err := api.Sites.Get(ctx.Args()[0], &*resourceGroup.URL)
	if err != nil {
		fatal(err.Error())
	}
	if err := api.Sites.Delete(*site.URL); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("%s site has been deleted.", *site.Name))
}

func sitesEnvCmd(c *CLI, ctx *cli.Context) {
	api := c.GetAPIClient(ctx)
	site := c.GetSite(ctx)
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
					Site:  site.URL,
					Key:   &parts[0],
					Value: &parts[1],
				}
				desiredEnvVars = append(desiredEnvVars, &envVar)
			}
		}
	}
	if !createMode {
		displayEnvVars, err = api.EnvVars.ListBySite(*site.URL)
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

func sitesUsersListCmd(c *CLI, ctx *cli.Context) {
	site := c.GetSite(ctx)
	users, err := site.GetUsers()
	if err != nil {
		fatal(err.Error())
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Username", "Role"})
	for i := range users {
		user := users[i]
		table.Append([]string{
			*user.Username,
			*user.Role,
		})
	}
	table.Render()
}

func sitesUsersAddCmd(c *CLI, ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Printf("Usage: %s sites users add [--role=dev] <email>\n", c.Name)
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	site := c.GetSite(ctx)
	email := ctx.Args()[0]
	if err := site.AddUser(email, ctx.String("role")); err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("added %q to %s", email, site.Name))
}
