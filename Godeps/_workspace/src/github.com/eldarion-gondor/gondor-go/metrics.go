package gondor

type MetricResource struct {
	client *Client
}

type MetricSeries struct {
	Columns []string `json:"columns"`
	Name    *string  `json:"name"`
	Points  [][]int  `json:"points"`
}

func (r *MetricResource) List(serviceURL string) ([]*MetricSeries, error) {
	url := r.client.buildBaseURL("metrics/")
	q := url.Query()
	q.Add("service", serviceURL)
	url.RawQuery = q.Encode()
	var res []*MetricSeries
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
