package v1alpha1

import (
	apiv1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GaiaNode is the Schema for the guestnodes API
// +genclient
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Runtime",type=string,JSONPath=`.spec.runtime`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.nodeStatus`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="ServiceStatus",type=string,priority=1,JSONPath=`.status.serviceStatus`
type GaiaNode struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GaiaNodeSpec   `json:"spec,omitempty"`
	Status GaiaNodeStatus `json:"status,omitempty"`
}

// GaiaNodeList contains a list of GuestNode
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GaiaNodeList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata,omitempty"`
	Items            []GaiaNode `json:"items"`
}

// GaiaNodeSpec defines the desired state of GuestNode
type GaiaNodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The name of the node in cluster
	Name string `json:"name"`

	// The host (domain) name of the node.
	Host string `json:"host,omitempty"`

	// Type of node, it must reference one of node types in the template.
	Type string `json:"type"`

	// The pvc name of the node.
	Pvc []Pvc `json:"pvc,omitempty"`

	// The tolerations of the node.
	Tolerations []apiv1.Toleration `json:"tolerations,omitempty"`

	// The label of the node.
	Labels []Labels `json:"labels,omitempty"`

	// Node deployment feature
	DeploymentFeature DeployFeature `json:"deploymentFeature,omitempty"`

	// The resources used by the node.
	Resource apiv1.ResourceRequirements `json:"resource,omitempty"`

	// The host IP which the user expect to run their node
	ExpectHost string `json:"expectHost,omitempty"`

	// Runtime, VM or Runc or Kata
	Runtime GuestNodeRuntime `json:"runtime,omitempty"`

	// Image used by the startup node.
	Image string `json:"image,omitempty"`

	// SidecarImage used by the startup vmnode.
	SidecarImage string `json:"sidecarimage,omitempty"`

	// The annotations at the node.
	Annotations map[string]string `json:"annotations,omitempty"`

	// The variables at the node level.
	Vars map[string]string `json:"vars,omitempty"`

	// The directory that the node mounts on.
	WorkPath string `json:"workPath,omitempty"`

	// The services running in the Guest Node, key of map define service type
	Services map[string]Service `json:"services,omitempty"`

	// The process running when befor node is deleted
	RemoveAction Process `json:"removeAction,omitempty"`

	// The node networkcards
	NetworkCards []NetworkCardConf `json:"networkcards,omitempty"`

	// The node portforwards
	PortForWards []PortForWardConf `json:"portforwards,omitempty"`

	//节点产生的文件
	Files []FileConf `json:"files,omitempty"`

	//节点映射到主机的端口
	Ports []PortConf `json:"ports,omitempty"`

	//node级别的dns记录描述  node level
	HostAliases []HostAlias `json:"hostAliases,omitempty"`

	//节点挂载目录或文件
	Volumes []VolumeConf `json:"volumes,omitempty"`
}

// GaiaNodeStatus defines the observed state of GuestNode
type GaiaNodeStatus struct {
	//Gaia wrapper can connect 11.254.0.1:443
	WrapperRegister string `json:"wrapperRegister,omitempty"`
	// IP 所在的宿主IP
	PhysicalIP string `json:"physicalIP,omitempty"`
	// Network Network
	Networks []NetworkInfo `json:"networks,omitempty"`
	// Storage 存储信息
	Storage string `json:"storage,omitempty"`
	// The state of this node.
	NodeState NodeState `json:"nodeStatus,omitempty"`
	// The services state at the node level.
	// The key of the map describes which service(type) and the value describes the state.
	SvcStates map[string]SvcState `json:"serviceStatus,omitempty"`

	PortForWardStates map[string]PortForWardState `json:"portforwardStatus,omitempty"`
}
