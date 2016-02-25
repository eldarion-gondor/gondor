package gondor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) Authenticate(username, password string) error {
	resp, err := http.PostForm(
		fmt.Sprintf("%s/oauth/token/", c.cfg.IdentityURL),
		url.Values{
			"grant_type": {"password"},
			"client_id":  {c.cfg.ID},
			"username":   {username},
			"password":   {password},
		},
	)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.New("authentication failed")
	}
	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if payload.Error != "" {
		return fmt.Errorf("authentication request failed: %q", payload.ErrorDescription)
	}
	c.cfg.Auth.Username = username
	c.cfg.Auth.AccessToken = payload.AccessToken
	c.cfg.Auth.RefreshToken = payload.RefreshToken
	if err := c.cfg.Persist(); err != nil {
		return err
	}
	return nil
}

func (c *Client) AuthenticateWithRefreshToken() error {
	resp, err := http.PostForm(
		fmt.Sprintf("%s/oauth/token/", c.cfg.IdentityURL),
		url.Values{
			"grant_type":    {"refresh_token"},
			"client_id":     {c.cfg.ID},
			"refresh_token": {c.cfg.Auth.RefreshToken},
		},
	)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.New("authentication failed")
	}
	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if payload.Error != "" {
		return fmt.Errorf("authentication request failed: %q", payload.ErrorDescription)
	}
	c.cfg.Auth.AccessToken = payload.AccessToken
	c.cfg.Auth.RefreshToken = payload.RefreshToken
	if err := c.cfg.Persist(); err != nil {
		return err
	}
	return nil
}

func (c *Client) RevokeAccess() error {
	resp, err := http.PostForm(
		fmt.Sprintf("%s/oauth/revoke_token/", c.cfg.IdentityURL),
		url.Values{
			"client_id": {c.cfg.ID},
			"token":     {c.cfg.Auth.RefreshToken},
		},
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unable to log out (%s)", resp.Status)
	}
	c.cfg.Auth.Username = ""
	c.cfg.Auth.AccessToken = ""
	c.cfg.Auth.RefreshToken = ""
	if err := c.cfg.Persist(); err != nil {
		return err
	}
	return nil
}
