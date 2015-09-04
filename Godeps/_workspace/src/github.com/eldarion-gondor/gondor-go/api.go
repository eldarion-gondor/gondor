package gondor

import (
	"fmt"
	"net/http"
	"net/url"
)

type ClientOptsPersister interface {
	Persist(*ClientOpts) error
}

type ClientOpts struct {
	ID          string
	BaseURL     string
	IdentityURL string
	Auth        struct {
		Username     string
		AccessToken  string
		RefreshToken string
	}
	Persister ClientOptsPersister
}

func (opts *ClientOpts) Persist() error {
	return opts.Persister.Persist(opts)
}

type Client struct {
	opts *ClientOpts

	httpClient *http.Client

	ResourceGroups *ResourceGroupResource
	Sites          *SiteResource
	Instances      *InstanceResource
	Releases       *ReleaseResource
	Services       *ServiceResource
	Builds         *BuildResource
	Deployments    *DeploymentResource
	HostNames      *HostNameResource
	KeyPairs       *KeyPairResource
	EnvVars        *EnvironmentVariableResource
	Logs           *LogResource
	Metrics        *MetricResource
}

func NewClient(opts *ClientOpts) *Client {
	c := &Client{
		opts:       opts,
		httpClient: http.DefaultClient,
	}
	c.attachResources()
	return c
}

func (c *Client) attachResources() {
	c.ResourceGroups = &ResourceGroupResource{client: c}
	c.Sites = &SiteResource{client: c}
	c.Instances = &InstanceResource{client: c}
	c.Releases = &ReleaseResource{client: c}
	c.Services = &ServiceResource{client: c}
	c.Builds = &BuildResource{client: c}
	c.Deployments = &DeploymentResource{client: c}
	c.HostNames = &HostNameResource{client: c}
	c.KeyPairs = &KeyPairResource{client: c}
	c.EnvVars = &EnvironmentVariableResource{client: c}
	c.Logs = &LogResource{client: c}
	c.Metrics = &MetricResource{client: c}
}

func (c *Client) buildBaseURL(endpoint string) *url.URL {
	url, err := url.Parse(c.opts.BaseURL)
	if err != nil {
		panic(fmt.Sprintf("bad base URL: %s", err.Error()))
	}
	url.Path = fmt.Sprintf("v2/%s", endpoint)
	return url
}
