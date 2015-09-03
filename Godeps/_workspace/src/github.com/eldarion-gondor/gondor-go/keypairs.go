package gondor

import (
	"errors"
	"fmt"
	"net/url"
)

type KeyPairResource struct {
	client *Client
}

type KeyPair struct {
	ResourceGroup *ResourceGroup `json:"resource_group,omitempty"`
	Name          string         `json:"name,omitempty"`
	Key           []byte         `json:"key,omitempty"`
	Certificate   []byte         `json:"certificate,omitempty"`

	URL string `json:"url,omitempty"`

	r *KeyPairResource
}

func (r *KeyPairResource) findOne(url *url.URL) (*KeyPair, error) {
	var res *KeyPair
	resp, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("keypair not found")
	}
	res.r = r
	return res, nil
}

func (r *KeyPairResource) GetByName(name string, resourceGroup *ResourceGroup) (*KeyPair, error) {
	url := r.client.buildBaseURL("keypairs/find/")
	q := url.Query()
	q.Set("name", name)
	if resourceGroup != nil {
		q.Set("resource_group", resourceGroup.URL)
	}
	url.RawQuery = q.Encode()
	return r.findOne(url)
}

func (r *KeyPairResource) List(resourceGroup *ResourceGroup) ([]*KeyPair, error) {
	url := r.client.buildBaseURL("keypairs/")
	q := url.Query()
	if resourceGroup != nil {
		q.Set("resource_group", resourceGroup.URL)
	}
	url.RawQuery = q.Encode()
	var res []*KeyPair
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *KeyPairResource) Create(keypair *KeyPair) error {
	url := r.client.buildBaseURL("keypairs/")
	_, err := r.client.Post(url, keypair, keypair)
	if err != nil {
		return err
	}
	return nil
}

func (r *KeyPairResource) Delete(keypair *KeyPair) error {
	if keypair.URL == "" {
		return errors.New("missing keypair URL")
	}
	u, _ := url.Parse(keypair.URL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}
