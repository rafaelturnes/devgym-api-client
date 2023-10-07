package kubeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/google/uuid"
)

func (c *KubeApiClient) CreateDeployment(nReplicas int, image string, ports []Ports, labels Labels) (*Deployment, error) {
	uuid := uuid.New().String()

	return c.CreateDeploymentUUID(uuid, nReplicas, image, ports, labels)
}

func (c *KubeApiClient) CreateDeploymentUUID(uuid string, nReplicas int, image string, ports []Ports, labels Labels) (*Deployment, error) {

	payload := Deployment{
		ID:       uuid,
		Image:    image,
		Ports:    ports,
		Replicas: nReplicas,
		Labels:   labels,
	}
	headers := map[string]string{"Content-Type": "application/json"}

	resp, err := c.request(http.MethodPost, "/deployments", headers, payload)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// cannot create the deployment
	if resp.StatusCode != 201 {
		code := strconv.Itoa(resp.StatusCode)
		re5xx := regexp.MustCompile(`5[0-9][0-9]`)
		switch code {
		case "400":
			return nil, NewErrBadRequest(resp)
		case "409":
			return nil, NewErrResponse(resp)
		case re5xx.FindString(code):
			return nil, NewErrResponse(resp)
		default:
			return nil, NewErrResponse(resp)
		}
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	err = json.Unmarshal(bs, &deployment)
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (c *KubeApiClient) GetDeployment(uuid string) (*Deployment, error) {
	endpoint := fmt.Sprintf("/deployments/%s", uuid)

	resp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		code := strconv.Itoa(resp.StatusCode)
		re5xx := regexp.MustCompile(`5[0-9][0-9]`)
		switch code {
		case "404":
			return nil, NewErrResponse(resp)
		case re5xx.FindString(code):
			return nil, NewErrResponse(resp)
		default:
			return nil, NewErrResponse(resp)
		}
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	err = json.Unmarshal(bs, &deployment)
	if err != nil {
		return nil, err
	}

	return &deployment, nil

}

func (c *KubeApiClient) DeleteDeployment(uuid string) error {

	endpoint := fmt.Sprintf("/deployments/%s", uuid)

	resp, err := c.request(http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Status Code 204 - success
	// Status Code 404 - deployment not found
	// Status code 5xx - API internal server error
	if resp.StatusCode != 204 {
		code := strconv.Itoa(resp.StatusCode)
		re5xx := regexp.MustCompile(`5[0-9][0-9]`)
		switch code {
		case "404":
			return NewErrResponse(resp)
		case re5xx.FindString(code):
			return NewErrResponse(resp)
		default:
			return NewErrResponse(resp)
		}
	}

	return nil
}
