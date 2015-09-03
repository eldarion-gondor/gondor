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
	resp, err := r.client.Session.Get(url.String(), nil, &res, nil)
	if err != nil {
		return nil, err
	}
	if resp.Status() == 404 {
		return nil, fmt.Errorf("instance not found")
	}
	err = respError(resp, nil)
	if err != nil {
		return nil, err
	}
	res.r = r
	return res, nil
}

func (r *InstanceResource) Create(instance *Instance) error {
	url := fmt.Sprintf("%s/v2/instances/", r.client.BaseURL)
	var errors ErrorList
	resp, err := r.client.Session.Post(url, instance, instance, &errors)
	if err != nil {
		return err
	}
	err = respError(resp, &errors)
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
	return r.findOne(url)
}

func (r *InstanceResource) Delete(instance *Instance) error {
	if instance.URL == "" {
		return errors.New("missing instance URL")
	}
	var errList ErrorList
	resp, err := r.client.Session.Delete(instance.URL, nil, &errList)
	if err != nil {
		return err
	}
	err = respError(resp, &errList)
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
	url := i.URL + "run/"
	var errList ErrorList
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
	resp, err := i.r.client.Session.Post(url, &up, &down, &errList)
	err = respError(resp, &errList)
	if err != nil {
		return "", err
	}
	return down.Endpoint, nil
}
