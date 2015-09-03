package gondor

import "net/url"

type EnvironmentVariableResource struct {
	client *Client
}

type EnvironmentVariable struct {
	Site     *Site     `json:"site,omitempty"`
	Instance *Instance `json:"instance,omitempty"`
	Service  *Service  `json:"service,omitempty"`
	Key      string    `json:"key,omitempty"`
	Value    string    `json:"value,omitempty"`

	URL string `json:"url,omitempty"`

	r *EnvironmentVariableResource
}

func (r *EnvironmentVariableResource) findMany(url *url.URL) ([]*EnvironmentVariable, error) {
	var res []*EnvironmentVariable
	resp, err := r.client.Session.Get(url.String(), nil, &res, nil)
	if err != nil {
		return nil, err
	}
	err = respError(resp, nil)
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
	var errors []ErrorList
	resp, err := r.client.Session.Post(url.String(), envVars, &envVars, &errors)
	if err != nil {
		return err
	}
	var resErrors ErrorList
	if len(errors) > 0 {
		resErrors = errors[0]
	}
	err = respError(resp, &resErrors)
	if err != nil {
		return err
	}
	return nil
}

func (r *EnvironmentVariableResource) ListBySite(site *Site) ([]*EnvironmentVariable, error) {
	url := r.client.buildBaseURL("envvars/")
	q := url.Query()
	q.Set("site", site.URL)
	url.RawQuery = q.Encode()
	return r.findMany(url)
}

func (r *EnvironmentVariableResource) ListByInstance(instance *Instance) ([]*EnvironmentVariable, error) {
	url := r.client.buildBaseURL("envvars/")
	q := url.Query()
	q.Set("instance", instance.URL)
	url.RawQuery = q.Encode()
	return r.findMany(url)
}

func (r *EnvironmentVariableResource) ListByService(service *Service) ([]*EnvironmentVariable, error) {
	url := r.client.buildBaseURL("envvars/")
	q := url.Query()
	q.Set("service", service.URL)
	url.RawQuery = q.Encode()
	return r.findMany(url)
}
