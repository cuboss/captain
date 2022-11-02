package monitoring

type Level int

const (
	LevelCluster = 1 << iota
	LevelNode
	LevelWorkload
	LevelPod
	LevelContainer
)

var MeteringLevelMap = map[string]int{
	"LevelCluster":   LevelCluster,
	"LevelNode":      LevelNode,
	"LevelWorkload":  LevelWorkload,
	"LevelPod":       LevelPod,
	"LevelContainer": LevelContainer,
}

type QueryOption interface {
	Apply(*QueryOptions)
}

type QueryOptions struct {
	Level Level

	NamespacedResourcesFilter string
	ResourceFilter            string

	NamespaceName string
	NodeName      string
	WorkloadKind  string
	WorkloadName  string
	PodName       string
	ContainerName string
}

func NewQueryOptions() *QueryOptions {
	return &QueryOptions{}
}

type ClusterOption struct{}

func (_ ClusterOption) Apply(o *QueryOptions) {
	o.Level = LevelCluster
}

type NodeOption struct {
	ResourceFilter string
	NodeName       string
}

func (no NodeOption) Apply(o *QueryOptions) {
	o.Level = LevelNode
	o.NodeName = no.NodeName
	o.ResourceFilter = no.ResourceFilter
}

type WorkloadOption struct {
	ResourceFilter string
	NamespaceName  string
	WorkloadKind   string
}

func (wo WorkloadOption) Apply(o *QueryOptions) {
	o.Level = LevelWorkload
	o.ResourceFilter = wo.ResourceFilter
	o.NamespaceName = wo.NamespaceName
	o.WorkloadKind = wo.WorkloadKind
}

type PodOption struct {
	NamespacedResourcesFilter string
	ResourceFilter            string

	NodeName      string
	NamespaceName string
	WorkloadKind  string
	WorkloadName  string
	PodName       string
}

func (po PodOption) Apply(o *QueryOptions) {
	o.Level = LevelPod
	o.NamespacedResourcesFilter = po.NamespacedResourcesFilter
	o.ResourceFilter = po.ResourceFilter
	o.NodeName = po.NodeName
	o.NamespaceName = po.NamespaceName
	o.WorkloadKind = po.WorkloadKind
	o.WorkloadName = po.WorkloadName
	o.PodName = po.PodName
}

type ContainerOption struct {
	ResourceFilter string
	NamespaceName  string
	PodName        string
	ContainerName  string
}

func (co ContainerOption) Apply(o *QueryOptions) {
	o.Level = LevelContainer
	o.ResourceFilter = co.ResourceFilter
	o.NamespaceName = co.NamespaceName
	o.PodName = co.PodName
	o.ContainerName = co.ContainerName
}
