package monitoring

const (
	MetricMeterPrefix = "meter_"
)

var ClusterMetrics = []string{
	"cluster_cpu_usage",
	"cluster_cpu_utilisation",
	"cluster_cpu_total",
	"cluster_memory_utilisation",
	"cluster_memory_available",
	"cluster_memory_total",
	"cluster_disk_size_usage",
	"cluster_disk_size_utilisation",
	"cluster_disk_size_capacity",
	"cluster_disk_size_available",
}

var NodeMetrics = []string{
	"node_cpu_utilisation",
	"node_cpu_usage",
	"node_cpu_total",
	"node_memory_utilisation",
	"node_memory_available",
	"node_memory_total",
	"node_memory_usage_wo_cache",
	"node_pod_count",
	"node_pod_quota",
}
