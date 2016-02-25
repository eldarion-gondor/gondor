package gondor

import (
	"fmt"
	"net/url"
)

type InstanceResource struct {
	client *Client
}

type Instance struct {
	Site   *string `json:"site,omitempty"`
	Label  *string `json:"label,omitempty"`
	Kind   *string `json:"kind,omitempty"`
	State  *string `json:"state,omitempty"`
	WebURL *string `json:"web_url,omitempty"`

	URL *string `json:"url,omitempty"`

	r *InstanceResource
}

func (r *InstanceResource) findOne(url *url.URL) (*Instance, error) {
	var res *Instance
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	res.r = r
	return res, nil
}

func (r *InstanceResource) Create(instance *Instance) error {
	url := r.client.buildBaseURL("instances/")
	_, err := r.client.Post(url, instance, instance)
	if err != nil {
		return err
	}
	return nil
}

func (r *InstanceResource) List(siteURL *string) ([]*Instance, error) {
	url := r.client.buildBaseURL("instances/")
	q := url.Query()
	if siteURL != nil {
		q.Set("site", *siteURL)
	}
	url.RawQuery = q.Encode()
	var res []*Instance
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].r = r
	}
	return res, nil
}

func (r *InstanceResource) GetFromURL(value string) (*Instance, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	return r.findOne(u)
}

func (r *InstanceResource) Get(siteURL string, label string) (*Instance, error) {
	url := r.client.buildBaseURL("instances/find/")
	q := url.Query()
	q.Set("site", siteURL)
	q.Set("label", label)
	url.RawQuery = q.Encode()
	instance, err := r.findOne(url)
	if _, ok := err.(ErrNotFound); ok {
		return instance, fmt.Errorf("instance %q was not found", label)
	}
	return instance, err
}

func (r *InstanceResource) Delete(instanceURL string) error {
	u, _ := url.Parse(instanceURL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}
