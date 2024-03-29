package conviva

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// responseError for Conviva error response
type responseError struct {
	Message string
	Request string
	Reason  string
	Code    int64
}

// httpClient interface to provide for Conviva API specific implementation
type httpClient interface {
	get(ctx context.Context, v interface{}, url string) (int, error)
}

type convivaHTTPClient struct {
	client   *http.Client
	username string
	password string
}

// newConvivaClient factory function for creating HTTPClientt
func newConvivaClient(client *http.Client, username string, password string) httpClient {
	return &convivaHTTPClient{
		client:   client,
		username: username,
		password: password,
	}
}

// Get method for Conviva API specific gets
func (c *convivaHTTPClient) get(ctx context.Context, v interface{}, url string) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req = req.WithContext(ctx)
	req.SetBasicAuth(c.username, c.password)
	res, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	if res.StatusCode != 200 {
		r := responseError{}
		if err = json.Unmarshal(body, &r); err == nil {
			return res.StatusCode, fmt.Errorf("%+v", r)
		}
		return res.StatusCode, fmt.Errorf("%+v", res)
	}
	err = json.Unmarshal(body, v)
	return res.StatusCode, err
}
