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

var WorkloadMetrics = []string{
	"workload_cpu_usage",
	"workload_memory_usage",
	"workload_memory_usage_wo_cache",

	"workload_deployment_replica",
	"workload_deployment_replica_available",
	"workload_statefulset_replica",
	"workload_statefulset_replica_available",
	"workload_daemonset_replica",
	"workload_daemonset_replica_available",
	"workload_deployment_unavailable_replicas_ratio",
	"workload_daemonset_unavailable_replicas_ratio",
	"workload_statefulset_unavailable_replicas_ratio",
}

var PodMetrics = []string{
	"pod_cpu_usage",
	"pod_memory_usage",
	"pod_memory_usage_wo_cache",
}

var ContainerMetrics = []string{
	"container_cpu_usage",
	"container_memory_usage",
	"container_memory_usage_wo_cache",
}
