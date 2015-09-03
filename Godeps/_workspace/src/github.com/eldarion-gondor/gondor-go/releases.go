package gondor

import (
	"errors"
	"net/url"
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
	url := r.client.buildBaseURL("releases/")
	release := Release{
		Instance: *instance,
	}
	_, err := r.client.Post(url, &release, &release)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *ReleaseResource) Delete(release *Release) error {
	if release.URL == "" {
		return errors.New("missing release URL")
	}
	u, _ := url.Parse(release.URL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}
