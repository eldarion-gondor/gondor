package gondorcli

import (
	"fmt"

	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"
)

func loginCmd(c *CLI, ctx *cli.Context) {
	if c.IsAuthenticated() {
		fatal(fmt.Sprintf("you are already logged in as %s. To log out run `%s logout`", c.Config.Identity.Username, c.Name))
	}
	api := c.GetAPIClient(ctx)
	// ask for username
	var username string
	fmt.Printf("Username: ")
	fmt.Scan(&username)
	// ask for password safely
	var password string
	password, err := speakeasy.Ask("Password: ")
	if err != nil {
		fatal(err.Error())
	}
	// authenticate user against identity
	if err := api.Authenticate(username, password); err != nil {
		fatal(err.Error())
	}
	// notify user
	success(fmt.Sprintf("logged in as %s", username))
}

func logoutCmd(c *CLI, ctx *cli.Context) {
	if !c.IsAuthenticated() {
		fatal("you are already logged out.")
	}
	api := c.GetAPIClient(ctx)
	if err := api.RevokeAccess(); err != nil {
		fatal(err.Error())
	}
	success("you have been logged out")
}
