package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"
)

func isAuthenticated() bool {
	return gcfg.loaded && gcfg.Auth.AccessToken != ""
}

func loginCmd(ctx *cli.Context) {
	if isAuthenticated() {
		fatal(fmt.Sprintf("you are already logged in as %s. To log out run `gondor logout`", gcfg.Auth.Username))
	}

	var username string
	fmt.Printf("Username: ")
	fmt.Scan(&username)

	var password string
	password, err := speakeasy.Ask("Password: ")

	resp, err := http.PostForm(
		"https://identity.gondor.io/oauth/token/",
		url.Values{
			"grant_type": {"password"},
			"client_id":  {"KtcICiPMAII8FAeArUoDB97zmjqltllyUDev8HOS"},
			"username":   {username},
			"password":   {password},
		},
	)
	if err != nil {
		fatal(err.Error())
	}
	if resp.StatusCode == 401 {
		fatal("authenication failed.")
	}
	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		AccessToken      string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		fatal(err.Error())
	}
	if payload.Error != "" {
		fatal(fmt.Sprintf("authentication request failed: %q", payload.ErrorDescription))
	}
	loginUser(username, payload.AccessToken)
}

func loginUser(username, accessToken string) {
	gcfg.Auth.Username = username
	gcfg.Auth.AccessToken = accessToken
	err := gcfg.Save()
	if err != nil {
		fatal(err.Error())
	}
	success(fmt.Sprintf("logged in as %s", username))
}

func logoutCmd(ctx *cli.Context) {
	if !isAuthenticated() {
		fatal("you are already logged out.")
	}
	resp, err := http.PostForm(
		"https://identity.gondor.io/oauth/revoke_token/",
		url.Values{
			"client_id": {"KtcICiPMAII8FAeArUoDB97zmjqltllyUDev8HOS"},
			"token":     {gcfg.Auth.AccessToken},
		},
	)
	if err != nil {
		fatal(err.Error())
	}
	if resp.StatusCode != 200 {
		fatal(fmt.Sprintf("unable to log out (%s)", resp.Status))
	}
	gcfg.Auth.Username = ""
	gcfg.Auth.AccessToken = ""
	if err := gcfg.Save(); err != nil {
		fatal(err.Error())
	}
	success("logged out")
}
