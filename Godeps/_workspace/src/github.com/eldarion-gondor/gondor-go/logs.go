package gondor

import "strconv"

type LogResource struct {
	client *Client
}

type LogRecord struct {
	Timestamp string `json:"@timestamp"`
	Message   string `json:"log"`
	Stream    string `json:"stream"`
	Tag       string `json:"tag"`
}

func (r *LogResource) ListByInstance(instance *Instance, lines int) ([]*LogRecord, error) {
	url := r.client.buildBaseURL("logs/")
	q := url.Query()
	q.Add("instance", instance.URL)
	q.Add("size", strconv.Itoa(lines))
	url.RawQuery = q.Encode()
	var res []*LogRecord
	resp, err := r.client.Session.Get(url.String(), nil, &res, nil)
	if err != nil {
		return nil, err
	}
	err = respError(resp, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *LogResource) ListByService(service *Service, lines int) ([]*LogRecord, error) {
	url := r.client.buildBaseURL("logs/")
	q := url.Query()
	q.Add("service", service.URL)
	q.Add("size", strconv.Itoa(lines))
	url.RawQuery = q.Encode()
	var res []*LogRecord
	resp, err := r.client.Session.Get(url.String(), nil, &res, nil)
	if err != nil {
		return nil, err
	}
	err = respError(resp, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}
