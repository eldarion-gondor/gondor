package gondor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) Authenticate(username, password string) (*ClientOpts, error) {
	resp, err := http.PostForm(
		"https://identity.gondor.io/oauth/token/",
		url.Values{
			"grant_type": {"password"},
			"client_id":  {c.opts.ID},
			"username":   {username},
			"password":   {password},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 {
		return nil, errors.New("authentication failed")
	}
	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Error != "" {
		return nil, fmt.Errorf("authentication request failed: %q", payload.ErrorDescription)
	}
	c.opts.Auth.Username = username
	c.opts.Auth.AccessToken = payload.AccessToken
	c.opts.Auth.RefreshToken = payload.RefreshToken
	return c.opts, nil
}

func (c *Client) RevokeAccess() (*ClientOpts, error) {
	resp, err := http.PostForm(
		"https://identity.gondor.io/oauth/revoke_token/",
		url.Values{
			"client_id": {c.opts.ID},
			"token":     {c.opts.Auth.RefreshToken},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to log out (%s)", resp.Status)
	}
	c.opts.Auth.Username = ""
	c.opts.Auth.AccessToken = ""
	c.opts.Auth.RefreshToken = ""
	return c.opts, nil
}
