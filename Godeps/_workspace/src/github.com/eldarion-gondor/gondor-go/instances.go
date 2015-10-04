package gondor

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type InstanceResource struct {
	client *Client
}

type Instance struct {
	Site     *Site     `json:"site,omitempty"`
	Label    string    `json:"label,omitempty"`
	Kind     string    `json:"kind,omitempty"`
	State    string    `json:"state,omitempty"`
	WebURL   string    `json:"web_url,omitempty"`
	Services []Service `json:"services,omitempty"`

	URL string `json:"url,omitempty"`

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

func (r *InstanceResource) GetFromURL(value string) (*Instance, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	return r.findOne(u)
}

func (r *InstanceResource) Get(site *Site, label string) (*Instance, error) {
	url := r.client.buildBaseURL("instances/find/")
	q := url.Query()
	q.Set("site", site.URL)
	q.Set("label", label)
	url.RawQuery = q.Encode()
	instance, err := r.findOne(url)
	if _, ok := err.(ErrNotFound); ok {
		return instance, fmt.Errorf("instance %q was not found", label)
	}
	return instance, err
}

func (r *InstanceResource) Delete(instance *Instance) error {
	if instance.URL == "" {
		return errors.New("missing instance URL")
	}
	u, _ := url.Parse(instance.URL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}

func (i *Instance) Load() error {
	newInstance, err := i.r.GetFromURL(i.URL)
	if err != nil {
		return err
	}
	*i = *newInstance
	return nil
}

func (i *Instance) Run(mode string, cmd []string) (string, error) {
	u, _ := url.Parse(i.URL + "run/")
	up := struct {
		Mode    string `json:"mode,omitempty"`
		Command string `json:"command,omitempty"`
	}{
		Mode:    mode,
		Command: strings.Join(cmd, " "),
	}
	down := struct {
		Endpoint string `json:"endpoint"`
	}{}
	_, err := i.r.client.Post(u, &up, &down)
	if err != nil {
		return "", err
	}
	return down.Endpoint, nil
}
