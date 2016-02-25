package gondor

type User struct {
	Username      string         `json:"username"`
	ResourceGroup *ResourceGroup `json:"resource_group"`
}

func (c *Client) AuthenticatedUser() (*User, error) {
	var res *User
	url := c.buildBaseURL("me/")
	_, err := c.Get(url, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
