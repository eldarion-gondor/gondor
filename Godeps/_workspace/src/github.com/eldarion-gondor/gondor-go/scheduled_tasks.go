package gondor

import (
	"errors"
	"net/url"
)

type ScheduledTaskResource struct {
	client *Client
}

type ScheduledTask struct {
	Instance *Instance `json:"instance,omitempty"`
	Name     string    `json:"name,omitempty"`
	Schedule string    `json:"schedule,omitempty"`
	Timezone string    `json:"timezone,omitempty"`
	Command  string    `json:"command,omitempty"`

	URL string `json:"url,omitempty"`

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

func (r *ScheduledTaskResource) List(instance *Instance) ([]*ScheduledTask, error) {
	v := url.Values{}
	if instance != nil {
		v.Add("instance", instance.URL)
	}
	url := r.client.buildBaseURL("scheduled_tasks/")
	q := url.Query()
	if instance != nil {
		q.Set("instance", instance.URL)
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

func (r *ScheduledTaskResource) DeleteByName(instance *Instance, name string) error {
	url := r.client.buildBaseURL("scheduled_tasks/find/")
	q := url.Query()
	q.Set("instance", instance.URL)
	q.Set("name", name)
	url.RawQuery = q.Encode()
	var res *ScheduledTask
	_, err := r.client.Get(url, &res)
	if err != nil {
		return err
	}
	return r.Delete(res)
}

func (r *ScheduledTaskResource) Delete(scheduledTask *ScheduledTask) error {
	if scheduledTask.URL == "" {
		return errors.New("scheduled task URL not defined")
	}
	u, _ := url.Parse(scheduledTask.URL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}
