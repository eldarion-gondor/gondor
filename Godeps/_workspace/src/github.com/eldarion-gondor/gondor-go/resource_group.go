package gondor

import (
	"errors"
	"fmt"
	"net/url"
)

type ResourceGroupResource struct {
	client *Client
}

type ResourceGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	URL string `json:"url"`

	r *ResourceGroupResource
}

func (r *ResourceGroupResource) findOne(url *url.URL) (*ResourceGroup, error) {
	var res *ResourceGroup
	resp, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("resource group not found")
	}
	res.r = r
	return res, nil
}

func (r *ResourceGroupResource) GetByName(name string) (*ResourceGroup, error) {
	url := r.client.buildBaseURL("resource_groups/find/")
	q := url.Query()
	q.Set("name", name)
	url.RawQuery = q.Encode()
	return r.findOne(url)
}

func (r *ResourceGroupResource) List() ([]*ResourceGroup, error) {
	url := r.client.buildBaseURL("resource_groups/")
	var res []*ResourceGroup
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ResourceGroupResource) Delete(resourceGroup *ResourceGroup) error {
	if resourceGroup.URL == "" {
		return errors.New("missing resource group URL")
	}
	u, _ := url.Parse(resourceGroup.URL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}