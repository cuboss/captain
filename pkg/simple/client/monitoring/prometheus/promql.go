package prometheus

import "captain/pkg/simple/client/monitoring"

var promQLTemplates = map[string]string{
	// cluster
	"cluster_cpu_utilisation":       ":node_cpu_utilisation:avg1m",
	"cluster_cpu_usage":             `round(:node_cpu_utilisation:avg1m * sum(node:node_num_cpu:sum), 0.001)`,
	"cluster_cpu_total":             "sum(node:node_num_cpu:sum)",
	"cluster_memory_utilisation":    ":node_memory_utilisation:",
	"cluster_memory_available":      "sum(node:node_memory_bytes_available:sum)",
	"cluster_memory_total":          "sum(node:node_memory_bytes_total:sum)",
	"cluster_disk_size_usage":       `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"} - node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
	"cluster_disk_size_utilisation": `cluster:disk_utilization:ratio`,
	"cluster_disk_size_capacity":    `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
	"cluster_disk_size_available":   `sum(max(node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
}

func makeExpr(metric string, opts monitoring.QueryOptions) string {
	tmpl := promQLTemplates[metric]
	switch opts.Level {
	case monitoring.LevelCluster:
		return tmpl
	default:
		return tmpl
	}
}
