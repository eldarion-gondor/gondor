package gondor

import "errors"

type DeploymentResource struct {
	client *Client
}

type Deployment struct {
	Service *string `json:"service,omitempty"`
	Build   *string `json:"build,omitempty"`

	URL *string `json:"url,omitempty"`

	r *DeploymentResource
}

func (r *DeploymentResource) Create(deployment *Deployment) error {
	url := r.client.buildBaseURL("deployments/")
	_, err := r.client.Post(url, deployment, deployment)
	if err != nil {
		return err
	}
	deployment.r = r
	return nil
}

func (d *Deployment) Wait() error {
	timeout := 60 * 15
	return WaitFor(timeout, func() (bool, error) {
		service, err := d.r.client.Services.GetFromURL(*d.Service)
		if err != nil {
			return false, err
		}
		switch *service.State {
		case "running":
			return true, nil
		case "deploying":
			return false, nil
		default:
			return false, errors.New("unknown instance state")
		}
	})
}
