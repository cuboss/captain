package component

import "time"

// 集群组件安装记录
type ClusterComponent struct {
	Id         int64     `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time;not null"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`

	CkeClusterId     string `json:"ckecluster_id" gorm:"column:cke_cluster_id;not null"`
	ClusterName      string `json:"cluster_name" gorm:"column:cluster_name;not null"`
	ComponentId      string `json:"component_id" gorm:"column:component_id;not null"`
	ComponentName    string `json:"component_name" gorm:"column:component_name;not null"`
	ComponentVersion string `json:"component_version" gorm:"column:component_version;not null"`

	Status string `json:"status" gorm:"column:status"`
	Values string `json:"values" gorm:"column:values"`

	ReleaseName string `json:"release_name" gorm:"column:release_name"`
	Namespace   string `json:"namespace" gorm:"column:namespace"`
}

type ClusterComponentResStatus struct {
	Kind      string
	Name      string
	Namespace string
	Ready     bool
	Message   string
}
