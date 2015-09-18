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

// SendRequest will build an HTTP request to send to the Gondor API.
func (c *Client) SendRequest(method string, url *url.URL, payload, result interface{}, attempts int) (*http.Response, error) {
	attempts++
	if attempts > 2 {
		return nil, errors.New("exceeded maximum retry limit")
	}
	var err error
	header := http.Header{}
	header.Add("Authorization", fmt.Sprintf("Bearer %s", c.cfg.Auth.AccessToken))
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
			switch resp.StatusCode {
			case 400:
				if err := json.Unmarshal(respBody, &errList); err != nil {
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
				return resp, apiError{errList: errList}
			case 401:
				c.AuthenticateWithRefreshToken()
				return c.SendRequest(method, url, payload, result, attempts)
			case 500:
				return resp, fmt.Errorf(
					"Internal Server Error\n%s",
					"Our staff has been notified of this error. Please try again later.",
				)
			case 502:
				return resp, fmt.Errorf(
					"Bad Gateway\n%s",
					"Our staff has been notified of this error. Please try again later.",
				)
			default:
				return resp, fmt.Errorf("unknown response: %d", resp.StatusCode)
			}
		}
	}
	return resp, nil
}

// Get issues an HTTP GET request
func (c *Client) Get(url *url.URL, result interface{}) (*http.Response, error) {
	return c.SendRequest("GET", url, nil, result, 0)
}

// Post issues an HTTP POST request
func (c *Client) Post(url *url.URL, payload, result interface{}) (*http.Response, error) {
	return c.SendRequest("POST", url, payload, result, 0)
}

// Put issues an HTTP PUT request
func (c *Client) Put(url *url.URL, payload, result interface{}) (*http.Response, error) {
	return c.SendRequest("PUT", url, payload, result, 0)
}

// Patch issues an HTTP PATCH request
func (c *Client) Patch(url *url.URL, payload, result interface{}) (*http.Response, error) {
	return c.SendRequest("PATCH", url, payload, result, 0)
}

// Delete issues an HTTP DELETE request
func (c *Client) Delete(url *url.URL, result interface{}) (*http.Response, error) {
	return c.SendRequest("DELETE", url, nil, result, 0)
}
