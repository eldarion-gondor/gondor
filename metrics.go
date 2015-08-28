package gondor

type MetricResource struct {
	client *Client
}

type MetricSeries struct {
	Columns []string `json:"columns"`
	Name    string   `json:"name"`
	Points  [][]int  `json:"points"`
}

func (r *MetricResource) List(service *Service) ([]*MetricSeries, error) {
	url := r.client.buildBaseURL("metrics/")
	q := url.Query()
	q.Add("service", service.URL)
	url.RawQuery = q.Encode()
	var res []*MetricSeries
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
