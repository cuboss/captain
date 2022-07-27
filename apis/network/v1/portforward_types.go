/*
Copyright 2021.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PortForwardSpec defines the desired state of PortForward
type PortForwardSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Port         int    `json:"port,omitempty"`
	Host         string `json:"host,omitempty"`
	PortType     string `json:"porttype,omitempty"`
	VpcName      string `json:"vpcname,omitempty"`
	VpcNamespace string `json:"vpcnamespace,omitempty"`
	TargetType   string `json:"targettype,omitempty"`
	TargetHost   string `json:"targethost,omitempty"`
	TargetPort   int    `json:"targetport"`
}

// PortForwardStatus defines the observed state of PortForward
type PortForwardStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Port         int      `json:"port"`
	Host         string   `json:"host"`
	PortType     string   `json:"porttype"`
	Nodes        []string `json:"nodes,omitempty"`
	VpcName      string   `json:"vpcname,omitempty"`
	VpcNamespace string   `json:"vpcnamespace,omitempty"`
	NetNS        string   `json:"netns,omitempty"`
	TargetType   string   `json:"targettype"`
	TargetHost   string   `json:"targethost"`
	TargetPort   int      `json:"targetport"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Host",type="string",JSONPath=".status.host",description="host for user to interactive"
//+kubebuilder:printcolumn:name="Port",type="string",JSONPath=".status.port",description="port for user to interactive"
//+kubebuilder:printcolumn:name="PortType",type="string",JSONPath=".status.porttype",description="port type for user to interactive"
//+kubebuilder:printcolumn:name="VpcNamespace",type="string",JSONPath=".status.vpcnamespace",description="target vpc namespace"
//+kubebuilder:printcolumn:name="VpcName",type="string",JSONPath=".status.vpcname",description="target vpcname"
//+kubebuilder:printcolumn:name="TargetHost",type="string",JSONPath=".status.targethost",description="targetip"
//+kubebuilder:printcolumn:name="TargetPort",type="string",JSONPath=".status.targetport",description="targetport"
//+kubebuilder:printcolumn:name="TargetType",type="string",JSONPath=".status.targettype",description="target port type for user to interactive"
//+kubebuilder:printcolumn:name="Nodes",type="string",JSONPath=".status.nodes",description="nodes"

// PortForward is the Schema for the portforwards API
type PortForward struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PortForwardSpec   `json:"spec,omitempty"`
	Status PortForwardStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PortForwardList contains a list of PortForward
type PortForwardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PortForward `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PortForward{}, &PortForwardList{})
}
