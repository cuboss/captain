package prometheus

import (
	"captain/pkg/simple/client/monitoring"
	"fmt"
	"strings"
)

var promQLTemplates = map[string]string{
	// cluster
	"cluster_cpu_utilisation":       ":node_cpu_utilisation:avg1m",
	"cluster_cpu_usage":             `round(:node_cpu_utilisation:avg1m * sum(node:node_num_cpu:sum), 0.001)`,
	"cluster_cpu_total":             "sum(node:node_num_cpu:sum)",
	"cluster_memory_utilisation":    ":node_memory_utilisation:",
	"cluster_memory_available":      "sum(node:node_memory_bytes_available:sum)",
	"cluster_memory_total":          "sum(node:node_memory_bytes_total:sum)",
	"cluster_memory_usage_wo_cache": "sum(node:node_memory_bytes_total:sum) - sum(node:node_memory_bytes_available:sum)",
	"cluster_disk_size_usage":       `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"} - node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
	"cluster_disk_size_utilisation": `cluster:disk_utilization:ratio`,
	"cluster_disk_size_capacity":    `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
	"cluster_disk_size_available":   `sum(max(node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,

	//node
	"node_cpu_utilisation":       "node:node_cpu_utilisation:avg1m{$1}",
	"node_cpu_usage":             `round(node:node_cpu_utilisation:avg1m{$1} * node:node_num_cpu:sum{$1}, 0.001)`,
	"node_cpu_total":             "node:node_num_cpu:sum{$1}",
	"node_memory_utilisation":    "node:node_memory_utilisation:{$1}",
	"node_memory_available":      "node:node_memory_bytes_available:sum{$1}",
	"node_memory_total":          "node:node_memory_bytes_total:sum{$1}",
	"node_memory_usage_wo_cache": "node:node_memory_bytes_total:sum{$1} - node:node_memory_bytes_available:sum{$1}",
	"node_pod_count":             `node:pod_count:sum{$1}`,
	"node_pod_quota":             `max(kube_node_status_capacity{resource="pods",$1}) by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)`,
}

func makeExpr(metric string, opts monitoring.QueryOptions) string {
	tmpl := promQLTemplates[metric]
	switch opts.Level {
	case monitoring.LevelCluster:
		return tmpl
	case monitoring.LevelNode:
		return makeNodeMetricExpr(tmpl, opts)
	default:
		return tmpl
	}
}

func makeNodeMetricExpr(tmpl string, o monitoring.QueryOptions) string {
	var nodeSelector string
	if o.NodeName != "" {
		nodeSelector = fmt.Sprintf(`node="%s"`, o.NodeName)
	} else {
		nodeSelector = fmt.Sprintf(`node=~"%s"`, o.ResourceFilter)
	}
	return strings.Replace(tmpl, "$1", nodeSelector, -1)
}
