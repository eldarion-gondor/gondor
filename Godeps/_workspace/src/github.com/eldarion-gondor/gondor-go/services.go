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
	resp, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("service not found")
	}
	res.r = r
	return res, nil
}

func (r *ServiceResource) Create(service *Service) error {
	url := r.client.buildBaseURL("services/")
	_, err := r.client.Post(url, service, service)
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
	u, _ := url.Parse(service.URL)
	service.URL = ""
	_, err := r.client.Patch(u, &service, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *ServiceResource) Delete(service *Service) error {
	if service.URL == "" {
		return errors.New("missing service URL")
	}
	u, _ := url.Parse(service.URL)
	_, err := r.client.Delete(u, nil)
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
	u, _ := url.Parse(s.URL)
	_, err := s.r.client.Patch(u, &desiredService, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SetReplicas(n int) error {
	desiredService := Service{
		DesiredReplicas: n,
	}
	u, _ := url.Parse(s.URL)
	_, err := s.r.client.Patch(u, &desiredService, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) DetachKeyPair() error {
	payload := struct {
		KeyPair *KeyPair `json:"keypair"`
	}{}
	u, _ := url.Parse(s.URL)
	_, err := s.r.client.Patch(u, &payload, nil)
	if err != nil {
		return err
	}
	return nil
}
