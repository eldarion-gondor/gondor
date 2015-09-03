package gondor

import (
	"errors"
	"fmt"
)

type ReleaseResource struct {
	client *Client
}

type Release struct {
	Instance Instance `json:"instance,omitempty"`
	Tag      string   `json:"tag,omitempty"`
	URL      string   `json:"url,omitempty"`

	r *ReleaseResource
}

func (r *ReleaseResource) Create(instance *Instance) (*Release, error) {
	url := fmt.Sprintf("%s/v2/releases/", r.client.BaseURL)
	release := Release{
		Instance: *instance,
	}
	var errors ErrorList
	resp, err := r.client.Session.Post(url, &release, &release, &errors)
	if err != nil {
		return nil, err
	}
	err = respError(resp, &errors)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *ReleaseResource) Delete(release *Release) error {
	if release.URL == "" {
		return errors.New("missing release URL")
	}
	var errList ErrorList
	resp, err := r.client.Session.Delete(release.URL, nil, &errList)
	if err != nil {
		return err
	}
	err = respError(resp, &errList)
	if err != nil {
		return err
	}
	return nil
}
