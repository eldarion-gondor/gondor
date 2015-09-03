package gondor

import (
	"errors"
	"fmt"
	"net/url"
)

type SiteResource struct {
	client *Client
}

type Site struct {
	Name          string         `json:"name,omitempty"`
	Key           string         `json:"key,omitempty"`
	ResourceGroup *ResourceGroup `json:"resource_group,omitempty"`
	Instances     []Instance     `json:"instances,omitempty"`
	Users         []struct {
		User struct {
			Username string `json:"username,omitempty"`
		} `json:"user,omitempty"`
		Role string `json:"role,omitempty"`
	} `json:"users,omitempty"`

	URL string `json:"url,omitempty"`

	r *SiteResource
}

func (r *SiteResource) Create(site *Site) error {
	url := fmt.Sprintf("%s/v2/sites/", r.client.BaseURL)
	var errors ErrorList
	resp, err := r.client.Session.Post(url, site, site, &errors)
	if err != nil {
		return err
	}
	err = respError(resp, &errors)
	if err != nil {
		return err
	}
	return nil
}

func (r *SiteResource) List(resourceGroup *ResourceGroup) ([]*Site, error) {
	url := r.client.buildBaseURL("sites/")
	q := url.Query()
	if resourceGroup != nil {
		q.Set("resource_group", resourceGroup.URL)
	}
	url.RawQuery = q.Encode()
	var res []*Site
	resp, err := r.client.Session.Get(url.String(), nil, &res, nil)
	if err != nil {
		return nil, err
	}
	err = respError(resp, nil)
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
	resp, err := r.client.Session.Get(url.String(), nil, &res, nil)
	if err != nil {
		return nil, err
	}
	if resp.Status() == 404 {
		return nil, fmt.Errorf("site not found")
	}
	err = respError(resp, nil)
	if err != nil {
		return nil, err
	}
	res.r = r
	return res, nil
}

func (r *SiteResource) Get(name string, resourceGroup *ResourceGroup) (*Site, error) {
	url := r.client.buildBaseURL("sites/find/")
	q := url.Query()
	q.Set("name", name)
	if resourceGroup != nil {
		q.Set("resource_group", resourceGroup.URL)
	}
	url.RawQuery = q.Encode()
	return r.findOne(url)
}

func (r *SiteResource) Delete(site *Site) error {
	if site.URL == "" {
		return errors.New("missing site URL")
	}
	var errList ErrorList
	resp, err := r.client.Session.Delete(site.URL, nil, &errList)
	if err != nil {
		return err
	}
	err = respError(resp, &errList)
	if err != nil {
		return err
	}
	return nil
}

func (site *Site) AddUser(email string, role string) error {
	url := fmt.Sprintf("%s/v2/site_users/", site.r.client.BaseURL)
	req := struct {
		Site  *Site  `json:"site,omitempty"`
		Email string `json:"email,omitempty"`
		Role  string `json:"role,omitempty"`
	}{
		Site:  site,
		Email: email,
		Role:  role,
	}
	var errList ErrorList
	resp, err := site.r.client.Session.Post(url, &req, nil, &errList)
	if err != nil {
		return err
	}
	err = respError(resp, &errList)
	if err != nil {
		return err
	}
	return nil
}
