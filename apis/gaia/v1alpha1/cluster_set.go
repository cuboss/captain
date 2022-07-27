package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GaiaSet is the Schema for the gaiaclustersets API
// +genclient
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GaiaSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GaiaSetSpec   `json:"spec,omitempty"`
	Status GaiaSetStatus `json:"status,omitempty"`
}

// GaiaSetList contains a list of GaiaSet
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GaiaSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GaiaSet `json:"items"`
}

// GaiaSetSpec defines the desired state of GaiaClusterSet
type GaiaSetSpec struct {
	// The template name for the cluster, it allows to select a specific configuration template.
	Template string `json:"template"`

	// runtime by VM/Runc
	Runtime GuestNodeRuntime `json:"runtime,omitempty"`

	// The name of the VPC network used by the cluster.
	VPC string `json:"vpc"`

	// Cluster deployment feature
	DeploymentFeature DeployFeature `json:"deploymentFeature,omitempty"`

	// Records of the hosts files added to each node in the cluster.
	HostAliases []HostAlias `json:"hostAliases,omitempty"` //集群级别的dns记录描述

	// The variables at the cluster level.
	Vars map[string]string `json:"vars,omitempty"`

	Nodes []string `json:"nodes"`
}

// GaiaSetStatus defines the observed state of GaiaSet
type GaiaSetStatus struct {
	// The key of the map describes which service(nodeType.servicetype) and the value describes the state.
	SvcStates map[string]SvcState `json:"serviceStatus,omitempty"`
}

//Equal 判断是否一致
func (gcss *GaiaSetStatus) Equal(status *GaiaSetStatus) bool {
	if gcss == status {
		return true
	}
	if gcss != nil && status != nil {
		if len(gcss.SvcStates) != len(status.SvcStates) {
			return false
		}
		if len(gcss.SvcStates) > 0 {
			for k, v := range gcss.SvcStates {
				if v1, ok := status.SvcStates[k]; !ok {
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
