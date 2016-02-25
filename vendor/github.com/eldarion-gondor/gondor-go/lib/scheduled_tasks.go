package gondor

import "net/url"

type ScheduledTaskResource struct {
	client *Client
}

type ScheduledTask struct {
	Instance *string `json:"instance,omitempty"`
	Name     *string `json:"name,omitempty"`
	Schedule *string `json:"schedule,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
	Command  *string `json:"command,omitempty"`

	URL *string `json:"url,omitempty"`

	r *ScheduledTaskResource
}

func (r *ScheduledTaskResource) Create(scheduledTask *ScheduledTask) error {
	url := r.client.buildBaseURL("scheduled_tasks/")
	_, err := r.client.Post(url, scheduledTask, scheduledTask)
	if err != nil {
		return err
	}
	return nil
}

func (r *ScheduledTaskResource) List(instanceURL *string) ([]*ScheduledTask, error) {
	url := r.client.buildBaseURL("scheduled_tasks/")
	q := url.Query()
	if instanceURL != nil {
		q.Set("instance", *instanceURL)
	}
	url.RawQuery = q.Encode()
	var res []*ScheduledTask
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].r = r
	}
	return res, nil
}

func (r *ScheduledTaskResource) DeleteByName(instanceURL string, name string) error {
	url := r.client.buildBaseURL("scheduled_tasks/find/")
	q := url.Query()
	q.Set("instance", instanceURL)
	q.Set("name", name)
	url.RawQuery = q.Encode()
	var res *ScheduledTask
	_, err := r.client.Get(url, &res)
	if err != nil {
		return err
	}
	return r.Delete(*res.URL)
}

func (r *ScheduledTaskResource) Delete(scheduledTaskURL string) error {
	u, _ := url.Parse(scheduledTaskURL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}
