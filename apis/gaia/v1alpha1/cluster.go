package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GaiaCluster is the Schema for the gaiaclusters API
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Template",type=string,JSONPath=`.spec.template`
// +kubebuilder:printcolumn:name="Vpc",type=string,JSONPath=`.spec.vpc`
// +kubebuilder:printcolumn:name="Progress",type=integer,JSONPath=`.status.clusterProgress`
// +kubebuilder:printcolumn:name="Runtime",type=string,JSONPath=`.spec.runtime`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="NodesProgress",type=string,priority=1,JSONPath=`.status.nodeProgress`
type GaiaCluster struct {
	meta_v1.TypeMeta   `json:",inline"`            //Kind, ApiVersion
	meta_v1.ObjectMeta `json:"metadata,omitempty"` //metadata.name, metadata.namespace...
	Spec               GaiaClusterSpec             `json:"spec"`
	Status             GaiaClusterStatus           `json:"status,omitempty"`
}

//GaiaClusterList contains a list of GuestCluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GaiaClusterList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []GaiaCluster `json:"items"`
}

//GaiaClusterSpec gaia cluster
type GaiaClusterSpec struct {
	// The template name for the cluster, it allows to select a specific configuration template.
	Template string `json:"template"`

	// runtime by VM/Runc
	Runtime GuestNodeRuntime `json:"runtime,omitempty"`

	// The name of the VPC network used by the cluster.
	VPC string `json:"vpc"`

	// ca auth ip range
	CaAuthIps []string `json:"caAuthIps,omitempty"`

	// service ip range
	ClusterIPRange string `json:"clusterIPRange,omitempty"`

	// pod cidr
	ClusterPodCidr string `json:"clusterPodCidr,omitempty"`

	Type string `json:"type,omitempty"`

	// coredns forward
	DnsForward string `json:"dnsForward,omitempty"`

	// Cluster deployment feature
	DeploymentFeature DeployFeature `json:"deploymentFeature,omitempty"`

	// Records of the hosts files added to each node in the cluster.
	HostAliases []HostAlias `json:"hostAliases,omitempty"` //集群级别的dns记录描述

	// The variables at the cluster level.
	Vars map[string]string `json:"vars,omitempty"`

	// Nodes defines the node and service info within the cluster.
	Nodes []GaiaNodeSpec `json:"nodes"`
}

// GaiaClusterStatus defines the observed state of GaiaCluster
type GaiaClusterStatus struct {
	//Installation progress
	ClusterProgress int `json:"clusterProgress"`
	// The installation progress of each node
	NodeStates map[string]int `json:"nodeProgress,omitempty"`
}

//Equal 判断是否一致
func (gcs *GaiaClusterStatus) Equal(status *GaiaClusterStatus) bool {
	if gcs == status {
		return true
	}
	if gcs.ClusterProgress != status.ClusterProgress {
		return false
	}
	if gcs != nil && status != nil {
		if len(gcs.NodeStates) != len(status.NodeStates) {
			return false
		}
		if len(gcs.NodeStates) > 0 {
			for k, v := range gcs.NodeStates {
				if v1, ok := status.NodeStates[k]; !ok {
					return false
				} else if v != v1 {
					return false
				}
			}
		}
		return true
	}
	return false
}
