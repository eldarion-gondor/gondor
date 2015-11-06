package gondor

import "errors"

type DeploymentResource struct {
	client *Client
}

type Deployment struct {
	Instance *string `json:"instance,omitempty"`
	Release  *string `json:"release,omitempty"`

	URL *string `json:"url,omitempty"`

	r *DeploymentResource
}

func (r *DeploymentResource) Create(deployment *Deployment) error {
	url := r.client.buildBaseURL("deployments/")
	_, err := r.client.Post(url, deployment, deployment)
	if err != nil {
		return err
	}
	return nil
}

func (d *Deployment) Wait() error {
	timeout := 60 * 15
	return WaitFor(timeout, func() (bool, error) {
		instance, err := d.r.client.Instances.GetFromURL(*d.Instance)
		if err != nil {
			return false, err
		}
		switch *instance.State {
		case "running":
			return true, nil
		case "deploying":
			return false, nil
		default:
			return false, errors.New("unknown instance state")
		}
	})
}
