package component

import (
	"time"
)

type ComponentBase struct {
	Type             string `json:"type"` // system or optional
	ComponentId      string `json:"component_id"`
	ComponentName    string `json:"component_name"`
	ComponentVersion string `json:"component_version"`
	ChartName        string `json:"chart_name"`
	ChartVersion     string `json:"chart_version"`
	ReleaseName      string `json:"release_name"`
}

// 集群组件安装记录
type ClusterComponent struct {
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`

	ComponentBase
	CkeClusterId string                 `json:"ckecluster_id"`
	ClusterName  string                 `json:"cluster_name"`
	Parameters   map[string]interface{} `json:"parameters"`
	Namespace    string                 `json:"namespace"`
	Status       string                 `json:"status"`
	DefultValues map[string]interface{} `json:"default_values"`
	ClusterComponentResources
}

type ClusterComponentResources struct {
	Resources []ClusterComponentResStatus `json:"resources"`
}
type ClusterComponentResStatus struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     bool   `json:"ready"`
	Message   string `json:"message"`
}

const (
	// TODO
	DefaultNamespace = "default"
)
