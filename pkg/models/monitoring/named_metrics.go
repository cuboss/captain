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
