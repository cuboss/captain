package monitoring

type Level int

const (
	LevelCluster = 1 << iota
	LevelNode
	LevelWorkload
	LevelPod
)

var MeteringLevelMap = map[string]int{
	"LevelCluster":  LevelCluster,
	"LevelNode":     LevelNode,
	"LevelWorkload": LevelWorkload,
	"LevelPod":      LevelPod,
}

type QueryOption interface {
	Apply(*QueryOptions)
}

type QueryOptions struct {
	Level Level

	NamespacedResourcesFilter string
	ResourceFilter            string
	NodeName                  string
	PodName                   string
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

type PodOption struct {
	NamespacedResourcesFilter string
	ResourceFilter            string
	PodName                   string
}

func (po PodOption) Apply(o *QueryOptions) {
	o.Level = LevelPod
	o.NamespacedResourcesFilter = po.NamespacedResourcesFilter
	o.ResourceFilter = po.ResourceFilter
	o.PodName = po.PodName
}
