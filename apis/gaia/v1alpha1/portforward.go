package v1alpha1

//PortForWardConf 定义端口
type PortForWardConf struct {
	Name         string `json:"name"` //名称
	Namespace    string `json:"namespace"`
	VPCName      string `json:"vpc_name,omitempty"` //VPC名称
	VPCNamespace string `json:"vpc_namespace"`
	Area         string `json:"area"` //区域
	TargetHost   string `json:"target_host,omitempty"`
	TargetPort   int    `json:"target_port"`
}
