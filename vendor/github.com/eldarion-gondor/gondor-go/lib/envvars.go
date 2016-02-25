package gondor

import "net/url"

type EnvironmentVariableResource struct {
	client *Client
}

type EnvironmentVariable struct {
	Site     *string `json:"site,omitempty"`
	Instance *string `json:"instance,omitempty"`
	Service  *string `json:"service,omitempty"`
	Key      *string `json:"key,omitempty"`
	Value    *string `json:"value,omitempty"`

	URL *string `json:"url,omitempty"`

	r *EnvironmentVariableResource
}

func (r *EnvironmentVariableResource) findMany(url *url.URL) ([]*EnvironmentVariable, error) {
	var res []*EnvironmentVariable
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].r = r
	}
	return res, nil
}

func (r *EnvironmentVariableResource) Create(envVars []*EnvironmentVariable) error {
	url := r.client.buildBaseURL("envvars/")
	_, err := r.client.Post(url, envVars, &envVars)
	if err != nil {
		return err
	}
	return nil
}

func (r *EnvironmentVariableResource) ListBySite(siteURL string) ([]*EnvironmentVariable, error) {
	url := r.client.buildBaseURL("envvars/")
	q := url.Query()
	q.Set("site", siteURL)
	url.RawQuery = q.Encode()
	return r.findMany(url)
}

func (r *EnvironmentVariableResource) ListByInstance(instanceURL string) ([]*EnvironmentVariable, error) {
	url := r.client.buildBaseURL("envvars/")
	q := url.Query()
	q.Set("instance", instanceURL)
	url.RawQuery = q.Encode()
	return r.findMany(url)
}

func (r *EnvironmentVariableResource) ListByService(serviceURL string) ([]*EnvironmentVariable, error) {
	url := r.client.buildBaseURL("envvars/")
	q := url.Query()
	q.Set("service", serviceURL)
	url.RawQuery = q.Encode()
	return r.findMany(url)
}
