package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/eldarion-gondor/gondor-go"
	"github.com/olekukonko/tablewriter"
)

func usersListCmd(ctx *cli.Context) {
	MustLoadSiteConfig()

	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	site := getSite(ctx, api)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Username", "Role"})
	for i := range site.Users {
		user := site.Users[i]
		table.Append([]string{
			user.User.Username,
			user.Role,
		})
	}
	table.Render()
}

func usersAddCmd(ctx *cli.Context) {
	usage := func(msg string) {
		fmt.Println("Usage: gondor users add [--role=dev] <email>")
		fatal(msg)
	}
	if len(ctx.Args()) == 0 {
		usage("too few arguments")
	}
	MustLoadSiteConfig()
	api := gondor.NewClient(ctx.GlobalString("api-url"), gcfg.Auth.AccessToken)
	site := getSite(ctx, api)
	if err := site.AddUser(ctx.Args()[0], ctx.String("role")); err != nil {
		fatal(err.Error())
	}
	fmt.Println("added")
}
