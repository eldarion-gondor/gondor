package gondor

import (
	"errors"
	"fmt"
	"net/url"
)

type ServiceResource struct {
	client *Client
}

type Service struct {
	Instance *Instance         `json:"instance,omitempty"`
	Name     string            `json:"name,omitempty"`
	Kind     string            `json:"kind,omitempty"`
	Replicas int               `json:"replicas,omitempty"`
	State    string            `json:"state,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	KeyPair  *KeyPair          `json:"keypair,omitempty"`

	// create only
	Version string `json:"version,omitempty"`

	// update only
	DesiredState    string `json:"desired_state,omitempty"`
	DesiredReplicas int    `json:"desired_replicas,omitempty"`

	URL string `json:"url,omitempty"`

	r *ServiceResource
}

func (r *ServiceResource) findOne(url *url.URL) (*Service, error) {
	var res *Service
	resp, err := r.client.Session.Get(url.String(), nil, &res, nil)
	if err != nil {
		return nil, err
	}
	if resp.Status() == 404 {
		return nil, fmt.Errorf("service not found")
	}
	err = respError(resp, nil)
	if err != nil {
		return nil, err
	}
	res.r = r
	return res, nil
}

func (r *ServiceResource) Create(service *Service) error {
	url := fmt.Sprintf("%s/v2/services/", r.client.BaseURL)
	var errors ErrorList
	resp, err := r.client.Session.Post(url, service, service, &errors)
	if err != nil {
		return err
	}
	err = respError(resp, &errors)
	if err != nil {
		return err
	}
	return nil
}

func (r *ServiceResource) GetFromURL(value string) (*Service, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	return r.findOne(u)
}

func (r *ServiceResource) Get(instance *Instance, name string) (*Service, error) {
	url := r.client.buildBaseURL("services/find/")
	q := url.Query()
	q.Set("instance", instance.URL)
	q.Set("name", name)
	url.RawQuery = q.Encode()
	return r.findOne(url)
}

func (r *ServiceResource) Update(service Service) error {
	url := service.URL
	service.URL = ""
	var errors ErrorList
	resp, err := r.client.Session.Patch(url, &service, nil, &errors)
	if err != nil {
		return err
	}
	err = respError(resp, &errors)
	if err != nil {
		return err
	}
	return nil
}

func (r *ServiceResource) Delete(service *Service) error {
	if service.URL == "" {
		return errors.New("missing service URL")
	}
	var errList ErrorList
	resp, err := r.client.Session.Delete(service.URL, nil, &errList)
	if err != nil {
		return err
	}
	err = respError(resp, &errList)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Restart() error {
	return s.SetState("restarted")
}

func (s *Service) SetState(state string) error {
	desiredService := Service{
		DesiredState: state,
	}
	var errors ErrorList
	resp, err := s.r.client.Session.Patch(s.URL, &desiredService, nil, &errors)
	if err != nil {
		return err
	}
	err = respError(resp, &errors)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SetReplicas(n int) error {
	desiredService := Service{
		DesiredReplicas: n,
	}
	var errors ErrorList
	resp, err := s.r.client.Session.Patch(s.URL, &desiredService, nil, &errors)
	if err != nil {
		return err
	}
	err = respError(resp, &errors)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) DetachKeyPair() error {
	payload := struct {
		KeyPair *KeyPair `json:"keypair"`
	}{}
	var errors ErrorList
	resp, err := s.r.client.Session.Patch(s.URL, &payload, nil, &errors)
	if err != nil {
		return err
	}
	err = respError(resp, &errors)
	if err != nil {
		return err
	}
	return nil
}
