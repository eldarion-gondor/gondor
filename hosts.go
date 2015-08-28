package gondor

import (
	"fmt"
	"net/url"
)

type HostNameResource struct {
	client *Client
}

type HostName struct {
	Instance *Instance `json:"instance,omitempty"`
	Host     string    `json:"host,omitempty"`

	URL string `json:"url,omitempty"`

	r *HostNameResource
}

func (r *HostNameResource) Create(hostName *HostName) error {
	url := fmt.Sprintf("%s/v2/hosts/", r.client.BaseURL)
	var errors ErrorList
	resp, err := r.client.Session.Post(url, hostName, hostName, &errors)
	if err != nil {
		return err
	}
	err = respError(resp, &errors)
	if err != nil {
		return err
	}
	return nil
}

func (r *HostNameResource) List(instance *Instance) ([]*HostName, error) {
	v := url.Values{}
	if instance != nil {
		v.Add("instance", instance.URL)
	}
	url := fmt.Sprintf("%s/v2/hosts/", r.client.BaseURL)
	if len(v) > 0 {
		url += "?" + v.Encode()
	}
	var res []*HostName
	resp, err := r.client.Session.Get(url, nil, &res, nil)
	if err != nil {
		return nil, err
	}
	switch resp.Status() {
	case 200:
		for i := range res {
			res[i].r = r
		}
		return res, nil
	default:
		return nil, fmt.Errorf("got unknown response: %d", resp.Status())
	}
}

func (r *HostNameResource) Delete(hostName *HostName) error {
	hostNames, err := r.List(hostName.Instance)
	if err != nil {
		return err
	}
	var foundHostName *HostName
	for i := range hostNames {
		foundHostName = hostNames[i]
		if hostName.Host == foundHostName.Host {
			break
		}
	}
	var errList ErrorList
	resp, err := r.client.Session.Delete(foundHostName.URL, nil, &errList)
	if err != nil {
		return err
	}
	err = respError(resp, &errList)
	if err != nil {
		return err
	}
	return nil
}
