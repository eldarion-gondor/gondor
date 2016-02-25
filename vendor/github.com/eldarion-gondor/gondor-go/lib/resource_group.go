package gondor

import (
	"fmt"
	"net/url"
)

type ResourceGroupResource struct {
	client *Client
}

type ResourceGroup struct {
	ID   *int    `json:"id"`
	Name *string `json:"name"`

	URL *string `json:"url"`

	r *ResourceGroupResource
}

func (r *ResourceGroupResource) findOne(url *url.URL) (*ResourceGroup, error) {
	var res *ResourceGroup
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	res.r = r
	return res, nil
}

func (r *ResourceGroupResource) GetFromURL(value string) (*ResourceGroup, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	return r.findOne(u)
}

func (r *ResourceGroupResource) GetByName(name string) (*ResourceGroup, error) {
	url := r.client.buildBaseURL("resource_groups/find/")
	q := url.Query()
	q.Set("name", name)
	url.RawQuery = q.Encode()
	resourceGroup, err := r.findOne(url)
	if _, ok := err.(ErrNotFound); ok {
		return resourceGroup, fmt.Errorf("resource group %q was not found", name)
	}
	return resourceGroup, err
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

func (r *ResourceGroupResource) Delete(resourceGroupURL string) error {
	u, _ := url.Parse(resourceGroupURL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}
