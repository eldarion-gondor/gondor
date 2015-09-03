package gondor

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jmcvetta/napping"
)

type Client struct {
	BaseURL     string
	AccessToken string

	Session *napping.Session

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

func NewClient(baseURL string, accessToken string) *Client {
	c := &Client{
		BaseURL:     baseURL,
		AccessToken: accessToken,
	}
	c.setupSession()
	c.attachResources()
	return c
}

func (c *Client) setupSession() {
	c.Session = &napping.Session{
		Client: http.DefaultClient,
		Header: &http.Header{},
	}
	c.Session.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
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
	url, err := url.Parse(c.BaseURL)
	if err != nil {
		panic(fmt.Sprintf("bad base URL: %s", err.Error()))
	}
	url.Path = fmt.Sprintf("v2/%s", endpoint)
	return url
}

func respError(resp *napping.Response, errs *ErrorList) error {
	switch resp.Status() {
	case 200, 201, 204:
		return nil
	case 401:
		return errors.New("unauthorized")
	case 400:
		if errs == nil {
			return errors.New("got 400 with no error list")
		}
		return apiError{errList: *errs}
	case 500:
		return fmt.Errorf("server error\nOur staff has been notified of this error. Please try again later.")
	case 502:
		return fmt.Errorf("bad gateway\nOur staff has been notified of this error. Please try again later.")
	default:
		return fmt.Errorf("got unknown response: %d", resp.Status())
	}
}
