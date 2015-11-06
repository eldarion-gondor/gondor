package gondor

import "net/url"

type ReleaseResource struct {
	client *Client
}

type Release struct {
	Instance *string `json:"instance,omitempty"`
	Tag      *string `json:"tag,omitempty"`

	URL *string `json:"url,omitempty"`

	r *ReleaseResource
}

func (r *ReleaseResource) Create(release *Release) error {
	url := r.client.buildBaseURL("releases/")
	_, err := r.client.Post(url, release, release)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReleaseResource) Delete(releaseURL string) error {
	u, _ := url.Parse(releaseURL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}
