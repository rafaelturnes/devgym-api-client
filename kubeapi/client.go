package kubeapi

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type KubeApiClient struct {
	url string
}

func NewClient(url string) *KubeApiClient {
	return &KubeApiClient{
		url: url,
	}
}

func (c *KubeApiClient) request(method string, endpoint string, header map[string]string, payload any) (*http.Response, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, c.url+endpoint, &buf)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient

	return client.Do(req)

}
