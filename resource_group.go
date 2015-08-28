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
	resp, err := r.client.Session.Get(url.String(), nil, &res, nil)
	if err != nil {
		return nil, err
	}
	if resp.Status() == 404 {
		return nil, fmt.Errorf("resource group not found")
	}
	err = respError(resp, nil)
	if err != nil {
		return nil, err
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
	url := fmt.Sprintf("%s/v2/resource_groups/", r.client.BaseURL)
	var res []*ResourceGroup
	resp, err := r.client.Session.Get(url, nil, &res, nil)
	if err != nil {
		return nil, err
	}
	switch resp.Status() {
	case 200:
		return res, nil
	default:
		return nil, fmt.Errorf("got unknown response: %d", resp.Status())
	}
}

func (r *ResourceGroupResource) Delete(resourceGroup *ResourceGroup) error {
	if resourceGroup.URL == "" {
		return errors.New("missing resource group URL")
	}
	var errList ErrorList
	resp, err := r.client.Session.Delete(resourceGroup.URL, nil, &errList)
	if err != nil {
		return err
	}
	err = respError(resp, &errList)
	if err != nil {
		return err
	}
	return nil
}
