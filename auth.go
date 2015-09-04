package main

import (
	"fmt"

	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"
)

func isAuthenticated() bool {
	return gcfg.loaded && gcfg.ClientOpts.Auth.AccessToken != ""
}

func loginCmd(ctx *cli.Context) {
	if isAuthenticated() {
		fatal(fmt.Sprintf("you are already logged in as %s. To log out run `gondor logout`", gcfg.ClientOpts.Auth.Username))
	}
	api := getAPIClient(ctx)
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

func logoutCmd(ctx *cli.Context) {
	if !isAuthenticated() {
		fatal("you are already logged out.")
	}
	api := getAPIClient(ctx)
	if err := api.RevokeAccess(); err != nil {
		fatal(err.Error())
	}
	success("you have been logged out")
}
