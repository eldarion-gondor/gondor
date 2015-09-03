package gondor

import "errors"

type DeploymentResource struct {
	client *Client
}

type Deployment struct {
	Instance *Instance `json:"instance,omitempty"`
	Release  Release   `json:"release,omitempty"`
	URL      string    `json:"url,omitempty"`

	r *DeploymentResource
}

func (r *DeploymentResource) Create(instance *Instance, release *Release) (*Deployment, error) {
	url := r.client.buildBaseURL("deployments/")
	deployment := Deployment{
		Instance: instance,
		Release:  *release,
	}
	_, err := r.client.Post(url, &deployment, &deployment)
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (d *Deployment) Wait() error {
	timeout := 60 * 15
	return WaitFor(timeout, func() (bool, error) {
		if err := d.Instance.Load(); err != nil {
			return false, err
		}
		switch d.Instance.State {
		case "running":
			return true, nil
		case "deploying":
			return false, nil
		default:
			return false, errors.New("unknown instance state")
		}
	})
}
