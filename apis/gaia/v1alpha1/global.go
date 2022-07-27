package v1alpha1

const (
	//ClusterLableName 集群名的标签名
	ClusterLableName = GroupName + "/cluster"
	//NodeLableName 主机名的标签名
	NodeLableName = GroupName + "/node"
)

//GuestNodeRuntime 节点运行类型
type GuestNodeRuntime string

const (
	// GuestNodeRuntimeVM vm
	GuestNodeRuntimeVM = "vm"
	// GuestNodeRuntimeRUNC runc
	GuestNodeRuntimeRUNC = "runc"
	// GuestNodeRuntimeKATA kata
	GuestNodeRuntimeKATA = "kata"
)

//NodeState 描述集群中各节点的当前运行状态
type NodeState string

//描述集群中各节点的当前状态
const (
	NodeSTOPPED  NodeState = "stopped"
	NodeSTARTING NodeState = "starting"
	NodeRUNNING  NodeState = "running"
	NodeSTOPPING NodeState = "stopping" //TODO: 需要确认是否需要这个状态
)

// NetworkInfo
type NetworkInfo struct {
	IP   string `json:"ipAddress,omitempty"`
	Mac  string `json:"macAddress,omitempty"`
	Name string `json:"name,omitempty"`
}

// PortForWardState
type PortForWardState struct {
	Name     string `json:"name,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	PortType string `json:"port_type,omitempty"`
}
