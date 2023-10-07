package kubeapi

import (
	"errors"
	"reflect"
	"testing"

	"slices"

	"github.com/google/uuid"
)

const url = "http://localhost:3000"

func TestCreateDeployment(t *testing.T) {
	client := NewClient(url)
	d := Deployment{
		Replicas: 1,
		Image:    "golang",
		Ports: []Ports{
			{Name: "http", Port: 80},
			{Name: "api", Port: 3000},
			{Name: "grpc", Port: 3001},
		},
		Labels: Labels{"app": "api", "env": "test"},
	}

	_, err := client.CreateDeployment(d.Replicas, d.Image, d.Ports, d.Labels)
	if err != nil {
		t.Errorf("deployment must be created, %v", err)
	}

}

func TestGetDeployment(t *testing.T) {
	client := NewClient(url)
	d := Deployment{
		Replicas: 1,
		Image:    "golang",
		Ports: []Ports{
			{Name: "http", Port: 80},
			{Name: "api", Port: 3000},
			{Name: "grpc", Port: 3001},
		},
		Labels: Labels{"app": "api", "env": "test"},
	}

	deploymentCreated, err := client.CreateDeployment(d.Replicas, d.Image, d.Ports, d.Labels)
	if err != nil {
		t.Errorf("deployment must be created, %v", err)
	}

	deploymentGet, err := client.GetDeployment(deploymentCreated.ID)
	if err != nil {
		t.Errorf("deployment must be found, %v", err)
	}

	if !reflect.DeepEqual(deploymentCreated, deploymentGet) {
		t.Errorf("deployment created and get are no equal")
	}

}

func TestGetDeployment_NotFound(t *testing.T) {
	client := NewClient(url)
	wantCode := 5
	fakeUUID := uuid.New().String()

	_, err := client.GetDeployment(fakeUUID)
	if err == nil {
		t.Errorf("deployment cannot be found, %v", err)
	}

	var errNotFound *ErrDefault
	if !errors.As(err, &errNotFound) {
		t.Errorf("error type must be an err response, %v", err)
	}

	errNotFound = err.(*ErrDefault)
	if errNotFound.Code != wantCode {
		t.Errorf("error code does not match, got: %d want: %d", errNotFound.Code, wantCode)
	}

}

func TestCreateDeployment_WrongFields(t *testing.T) {
	client := NewClient(url)

	cases := map[string]struct {
		deployment   Deployment
		code         int
		failedFields []string
	}{
		"empty deployment": {
			deployment:   Deployment{},
			code:         1032,
			failedFields: []string{"replicas", "image", "ports"},
		},
		"no ports": {
			deployment: Deployment{
				Replicas: 1,
				Image:    "golang",
				Ports:    []Ports{},
			},
			code:         1032,
			failedFields: []string{"ports"},
		},
		"ports out of range": {
			deployment: Deployment{
				Replicas: 1,
				Image:    "golang",
				Ports: []Ports{
					{Name: "http", Port: 80},
					{Name: "api", Port: 65600},
					{Name: "healthz", Port: 65600},
				},
			},
			code:         3020,
			failedFields: []string{},
		},
		"no image": {
			deployment: Deployment{
				Replicas: 1,
				Image:    "",
				Ports: []Ports{
					{Name: "http", Port: 80},
				},
			},
			code:         1032,
			failedFields: []string{"image"},
		},
		"no replicas": {
			deployment: Deployment{
				Replicas: 0,
				Image:    "golang",
				Ports: []Ports{
					{Name: "http", Port: 80},
				},
			},
			code:         1032,
			failedFields: []string{"replicas"},
		},
		// "negative replicas": {
		// 	deployment: Deployment{
		// 		Replicas: -5,
		// 		Image:    "golang",
		// 		Ports: []Ports{
		// 			{Name: "http", Port: 80},
		// 		},
		// 	},
		// 	code:         1032,
		// 	failedFields: []string{},
		// },
	}

	for test, v := range cases {
		t.Run(test, func(t *testing.T) {
			_, err := client.CreateDeployment(v.deployment.Replicas, v.deployment.Image, v.deployment.Ports, v.deployment.Labels)
			if err == nil {
				t.Errorf("%s must fail, %v", test, err)
				return
			}
			var badRequest *BadRequest
			if !errors.As(err, &badRequest) {
				t.Errorf("%s must be a bad request, %v", test, err)
				return
			}

			badRequest = err.(*BadRequest)
			if v.code != badRequest.Code {
				t.Errorf("%s error code does not match, got: %d want: %d", test, badRequest.Code, v.code)
				return
			}

			gotFailedFields := err.(*BadRequest).Extras.FailedFields
			if !slices.Equal[[]string](v.failedFields, gotFailedFields) {
				t.Errorf("%s must match the failed fields, got: %+v want: %+v", test, gotFailedFields, v.failedFields)
				return
			}

		})
	}

}

func TestCreateDeploymentUUID_Duplicated(t *testing.T) {
	uuid := uuid.New().String()
	cases := map[string]struct {
		deployment Deployment
		code       int
		httpcode   int
	}{
		"duplicated deployment": {
			deployment: Deployment{
				ID:       uuid,
				Replicas: 1,
				Image:    "golang",
				Labels:   Labels{"app": "api"},
				Ports:    []Ports{{Name: "http", Port: 3000}},
			},
			code:     5000,
			httpcode: 409,
		},
	}

	client := NewClient(url)

	for test, v := range cases {
		t.Run(test, func(t *testing.T) {

			_, err := client.CreateDeploymentUUID(v.deployment.ID, v.deployment.Replicas, v.deployment.Image, v.deployment.Ports, v.deployment.Labels)
			if err != nil {
				t.Errorf("first deployment must be created, %v", err)
			}

			_, err = client.CreateDeploymentUUID(v.deployment.ID, v.deployment.Replicas, v.deployment.Image, v.deployment.Ports, v.deployment.Labels)
			if err == nil {
				t.Errorf("second deployment must not be created, %v", err)
			}

			var statusConflict *ErrDefault
			if !errors.As(err, &statusConflict) {
				t.Errorf("Duplicated deployment must be a default err response, %v", err)
			}

			statusConflict = err.(*ErrDefault)
			if v.httpcode != statusConflict.HttpCode {
				t.Errorf("%s http code does not match, got: %d want: %d", test, statusConflict.HttpCode, v.httpcode)
				return
			}
		})
	}
}

func TestDeleteDeployment(t *testing.T) {
	client := NewClient(url)
	d := Deployment{
		Replicas: 1,
		Image:    "golang",
		Ports: []Ports{
			{Name: "http", Port: 80},
			{Name: "api", Port: 3000},
			{Name: "grpc", Port: 3001},
		},
		Labels: Labels{"app": "api", "env": "test"},
	}

	dCreated, err := client.CreateDeployment(d.Replicas, d.Image, d.Ports, d.Labels)
	if err != nil {
		t.Errorf("deployment must be created, %v", err)
	}

	err = client.DeleteDeployment(dCreated.ID)
	if err != nil {
		t.Errorf("deployment must be deleted, %v", err)
	}
}

func TestDeleteDeployment_NotFound(t *testing.T) {
	client := NewClient(url)
	fakeUUID := uuid.New().String()
	wantCode := 5
	err := client.DeleteDeployment(fakeUUID)
	if err == nil {
		t.Errorf("deployment must fail, uuid: %s", fakeUUID)
	}
	var notFoundErr *ErrDefault
	if !errors.As(err, &notFoundErr) {
		t.Errorf("error type must be an err response, %v", err)
	}

	notFoundErr = err.(*ErrDefault)
	if wantCode != notFoundErr.Code {
		t.Errorf("error code does not match, got: %d want: %d", notFoundErr.Code, wantCode)
	}
}
