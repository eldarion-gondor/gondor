package gondor

import (
	"fmt"
	"net/url"
)

type SiteResource struct {
	client *Client
}

type Site struct {
	Name          *string `json:"name,omitempty"`
	Key           *string `json:"key,omitempty"`
	ResourceGroup *string `json:"resource_group,omitempty"`

	URL *string `json:"url,omitempty"`

	r *SiteResource
}

type SiteUser struct {
	Site     *string `json:"site,omitempty"`
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Role     *string `json:"role,omitempty"`

	r *SiteResource
}

func (r *SiteResource) Create(site *Site) error {
	url := r.client.buildBaseURL("sites/")
	_, err := r.client.Post(url, site, site)
	if err != nil {
		return err
	}
	return nil
}

func (r *SiteResource) List(resourceGroupURL *string) ([]*Site, error) {
	url := r.client.buildBaseURL("sites/")
	q := url.Query()
	if resourceGroupURL != nil {
		q.Set("resource_group", *resourceGroupURL)
	}
	url.RawQuery = q.Encode()
	var res []*Site
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].r = r
	}
	return res, nil
}

func (r *SiteResource) findOne(url *url.URL) (*Site, error) {
	var res *Site
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	res.r = r
	return res, nil
}

func (r *SiteResource) Get(name string, resourceGroupURL *string) (*Site, error) {
	url := r.client.buildBaseURL("sites/find/")
	q := url.Query()
	q.Set("name", name)
	if resourceGroupURL != nil {
		q.Set("resource_group", *resourceGroupURL)
	}
	url.RawQuery = q.Encode()
	site, err := r.findOne(url)
	if _, ok := err.(ErrNotFound); ok {
		identifier := name
		if resourceGroupURL != nil {
			resourceGroup, err := r.client.ResourceGroups.GetFromURL(*resourceGroupURL)
			if err == nil {
				identifier = fmt.Sprintf("%s/%s", *resourceGroup.Name, name)
			}
		}
		return site, fmt.Errorf("site %q was not found", identifier)
	}
	return site, err
}

func (r *SiteResource) Delete(siteURL string) error {
	u, _ := url.Parse(siteURL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}

func (site *Site) AddUser(email string, role string) error {
	url := site.r.client.buildBaseURL("site_users/")
	req := &SiteUser{
		Site:  site.URL,
		Email: &email,
		Role:  &role,
	}
	_, err := site.r.client.Post(url, &req, nil)
	if err != nil {
		return err
	}
	return nil
}

func (site *Site) GetUsers() ([]*SiteUser, error) {
	url := site.r.client.buildBaseURL("site_users/")
	q := url.Query()
	q.Set("site", *site.URL)
	url.RawQuery = q.Encode()
	var res []*SiteUser
	_, err := site.r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].r = site.r
	}
	return res, nil
}
