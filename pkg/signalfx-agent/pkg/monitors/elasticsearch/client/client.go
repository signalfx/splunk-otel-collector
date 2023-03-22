package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ESClient struct {
	Scheme     string
	Host       string
	Port       string
	HTTPClient *http.Client
}

// Fetches a JSON response and puts it into an object
func (c *ESClient) FetchJSON(url string, obj interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("could not get url %s: %v", url, err)
	}

	res, err := c.HTTPClient.Do(req)

	if err != nil {
		return fmt.Errorf("could not get url %s: %v", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("received status code that's not 200: %s, url: %s", res.Status, url)
	}

	err = json.NewDecoder(res.Body).Decode(obj)

	if err != nil {
		return fmt.Errorf("could not get url %s: %v", url, err)
	}

	return nil
}
