package gondor

import (
	"fmt"
	"net/http"
	"net/url"
)

type ConfigPersister interface {
	Persist(*Config) error
}

type Config struct {
	ID          string
	BaseURL     string
	IdentityURL string
	Auth        struct {
		Username     string
		AccessToken  string
		RefreshToken string
	}
	Persister ConfigPersister
}

func (cfg *Config) Persist() error {
	return cfg.Persister.Persist(cfg)
}

type Client struct {
	cfg *Config

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

	logHTTP bool
}

func NewClient(cfg *Config, httpClient *http.Client) *Client {
	c := &Client{
		cfg:        cfg,
		httpClient: httpClient,
	}
	c.attachResources()
	return c
}

func (c *Client) EnableHTTPLogging(value bool) {
	c.logHTTP = value
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
	url, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		panic(fmt.Sprintf("bad base URL: %s", err.Error()))
	}
	url.Path = fmt.Sprintf("v2/%s", endpoint)
	return url
}
