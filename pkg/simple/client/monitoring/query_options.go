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
}

func NewQueryOptions() *QueryOptions {
	return &QueryOptions{}
}

type ClusterOption struct{}

func (_ ClusterOption) Apply(o *QueryOptions) {
	o.Level = LevelCluster
}

type PodOption struct {
	NamespacedResourcesFilter string
}

func (po PodOption) Apply(o *QueryOptions) {
	o.Level = LevelPod
	o.NamespacedResourcesFilter = po.NamespacedResourcesFilter
}
