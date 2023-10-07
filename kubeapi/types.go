package kubeapi

import (
	"time"
)

type Deployment struct {
	ID       string    `json:"id"`
	Replicas int       `json:"replicas"`
	Image    string    `json:"image"`
	Labels   Labels    `json:"labels"`
	Ports    []Ports   `json:"ports"`
	CreateAt time.Time `json:"createAt"`
}

type Ports struct {
	Name string `json:""`
	Port int    `json:""`
}

type Labels map[string]string
