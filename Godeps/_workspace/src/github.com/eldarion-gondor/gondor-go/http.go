package gondor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (c *Client) SendRequest(method string, url *url.URL, payload, result interface{}) (*http.Response, error) {
	var err error
	header := http.Header{}
	header.Add("Authorization", fmt.Sprintf("Bearer %s", c.opts.Auth.AccessToken))
	var body io.Reader
	if payload != nil {
		var b []byte
		b, err = json.Marshal(&payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(b)
		header.Add("Content-Type", "application/json")
	}
	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}
	header.Add("Accept", "application/json")
	req.Header = header
	var errList ErrorList
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if string(respBody) != "" {
		if resp.StatusCode < 300 && result != nil {
			if err := json.Unmarshal(respBody, result); err != nil {
				return resp, err
			}
		}
		if resp.StatusCode >= 400 {
			if err := json.Unmarshal(respBody, errList); err != nil {
				if verr, ok := err.(*json.UnmarshalTypeError); ok {
					if verr.Value == "array" {
						var errLofL []ErrorList
						err = json.Unmarshal(respBody, errLofL)
						if err == nil {
							if len(errLofL) > 0 {
								errList = errLofL[0]
							}
						}
					}
				}
				if err != nil {
					return resp, err
				}
			}
			switch resp.StatusCode {
			case 400:
				return resp, apiError{errList: errList}
			case 401:
				return resp, errors.New("unauthorized")
			case 500:
				return resp, fmt.Errorf("server error\nOur staff has been notified of this error. Please try again later.")
			case 502:
				return resp, fmt.Errorf("bad gateway\nOur staff has been notified of this error. Please try again later.")
			default:
				return resp, fmt.Errorf("unknown response: %d", resp.StatusCode)
			}
		}
	}
	return resp, nil
}

func (c *Client) Get(url *url.URL, result interface{}) (*http.Response, error) {
	return c.SendRequest("GET", url, nil, result)
}

func (c *Client) Post(url *url.URL, payload, result interface{}) (*http.Response, error) {
	return c.SendRequest("POST", url, payload, result)
}

func (c *Client) Put(url *url.URL, payload, result interface{}) (*http.Response, error) {
	return c.SendRequest("PUT", url, payload, result)
}

func (c *Client) Patch(url *url.URL, payload, result interface{}) (*http.Response, error) {
	return c.SendRequest("PATCH", url, payload, result)
}

func (c *Client) Delete(url *url.URL, result interface{}) (*http.Response, error) {
	return c.SendRequest("DELETE", url, nil, result)
}
