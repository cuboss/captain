package v1alpha1

//SuccessLevel 服务的完成影响范围
type SuccessLevel string

//描述集群中各服务的完成影响范围
const (
	SuccessNode    SuccessLevel = "node"
	SuccessCluster SuccessLevel = "cluster"
)

//SvcState 描述集群中各种服务的当前运行状态
type SvcState string

//描述集群中各种服务的当前状态
const (
	SvcSTOPPED  SvcState = "stop"
	SvcSTARTING SvcState = "starting"
	SvcRUNNING  SvcState = "running"
	SvcFINISHED SvcState = "finished"
	SvcError    SvcState = "error"
)

//ExpectState 描述集群中各种服务的期望状态
type ExpectState string

//描述集群中各种服务的期望状态
const (
	ExpectRUNNING  ExpectState = "running"
	ExpectFINISHED ExpectState = "finished"
)

// Process 定义GuestNode内一个业务所需执行的进程信息
type Process struct {
	// 进程标识
	ID string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// command want to run.
	Cmd string `json:"cmd"`

	// Env environments of process
	Envs []string `json:"envs,omitempty"`

	// Args arguments of process
	Args []string `json:"args,omitempty"`

	// 进程的标准输出文件
	Stdout string `protobuf:"bytes,6,opt,name=stdout,proto3" json:"stdout,omitempty"`
	// 进程的错误输出文件
	Stderr string `protobuf:"bytes,7,opt,name=stderr,proto3" json:"stderr,omitempty"`
	// 进程的输出最大大小
	MaxOutSize int64 `protobuf:"varint,8,opt,name=maxOutSize,proto3" json:"maxOutSize,omitempty"`
	// 进程输出的文件数
	MaxOutCount int32 `protobuf:"varint,9,opt,name=maxOutCount,proto3" json:"maxOutCount,omitempty"`
}

// Service 定义GuestNode内运行的业务的相关信息
type Service struct {
	// SuccessLevel 服务成功的层次(Cluster/Node)
	SuccessLevel SuccessLevel `json:"successLevel,omitempty"`

	// RunType 服务期望的状态
	RunType ExpectState `json:"runType,omitempty"`

	// Dependence 服务的依赖集
	Dependence []string `json:"dependence,omitempty"`

	// process info of service.
	Srv Process `json:"srv"`

	// The initialization process that runs before the service.
	Init Process `json:"init,omitempty"`

	// The checker that runs after the service.
	Check Process `json:"check,omitempty"`

	Unique bool `json:"unique,omitempty"`
}
