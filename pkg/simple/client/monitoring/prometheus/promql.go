package prometheus

import (
	"fmt"
	"strings"

	"captain/pkg/simple/client/monitoring"
)

const (
	StatefulSet = "StatefulSet"
	DaemonSet   = "DaemonSet"
	Deployment  = "Deployment"
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

	//  workload
	"workload_cpu_usage":             `round(namespace:workload_cpu_usage:sum{$1}, 0.001)`,
	"workload_memory_usage":          `namespace:workload_memory_usage:sum{$1}`,
	"workload_memory_usage_wo_cache": `namespace:workload_memory_usage_wo_cache:sum{$1}`,
	"workload_net_bytes_transmitted": `namespace:workload_net_bytes_transmitted:sum_irate{$1}`,
	"workload_net_bytes_received":    `namespace:workload_net_bytes_received:sum_irate{$1}`,

	"workload_deployment_replica":                     `label_join(sum (label_join(label_replace(kube_deployment_spec_replicas{$2}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_deployment_replica_available":           `label_join(sum (label_join(label_replace(kube_deployment_status_replicas_available{$2}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_statefulset_replica":                    `label_join(sum (label_join(label_replace(kube_statefulset_replicas{$2}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_statefulset_replica_available":          `label_join(sum (label_join(label_replace(kube_statefulset_status_replicas_current{$2}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_daemonset_replica":                      `label_join(sum (label_join(label_replace(kube_daemonset_status_desired_number_scheduled{$2}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_daemonset_replica_available":            `label_join(sum (label_join(label_replace(kube_daemonset_status_number_available{$2}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_deployment_unavailable_replicas_ratio":  `namespace:deployment_unavailable_replicas:ratio{$1}`,
	"workload_daemonset_unavailable_replicas_ratio":   `namespace:daemonset_unavailable_replicas:ratio{$1}`,
	"workload_statefulset_unavailable_replicas_ratio": `namespace:statefulset_unavailable_replicas:ratio{$1}`,

	// pod
	"pod_cpu_usage":             `round(sum by (namespace, pod) (irate(container_cpu_usage_seconds_total{job="kubelet", pod!="", image!=""}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}, 0.001)`,
	"pod_memory_usage":          `sum by (namespace, pod) (container_memory_usage_bytes{job="kubelet", pod!="", image!=""}) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}`,
	"pod_memory_usage_wo_cache": `sum by (namespace, pod) (container_memory_working_set_bytes{job="kubelet", pod!="", image!=""}) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}`,
	"pod_net_bytes_transmitted": `sum by (namespace, pod) (irate(container_network_transmit_bytes_total{pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}`,
	"pod_net_bytes_received":    `sum by (namespace, pod) (irate(container_network_receive_bytes_total{pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}`,

	// container
	"container_cpu_usage":             `round(sum by (namespace, pod, container) (irate(container_cpu_usage_seconds_total{job="kubelet", container!="POD", container!="", image!="", $1}[5m])), 0.001)`,
	"container_memory_usage":          `sum by (namespace, pod, container) (container_memory_usage_bytes{job="kubelet", container!="POD", container!="", image!="", $1})`,
	"container_memory_usage_wo_cache": `sum by (namespace, pod, container) (container_memory_working_set_bytes{job="kubelet", container!="POD", container!="", image!="", $1})`,
}

func makeExpr(metric string, opts monitoring.QueryOptions) string {
	tmpl := promQLTemplates[metric]
	switch opts.Level {
	case monitoring.LevelCluster:
		return tmpl
	case monitoring.LevelNode:
		return makeNodeMetricExpr(tmpl, opts)
	case monitoring.LevelWorkload:
		return makeWorkloadMetricExpr(metric, tmpl, opts)
	case monitoring.LevelPod:
		return makePodMetricExpr(tmpl, opts)
	case monitoring.LevelContainer:
		return makeContainerMetricExpr(tmpl, opts)
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

func makeWorkloadMetricExpr(metric, tmpl string, o monitoring.QueryOptions) string {
	var kindSelector, workloadSelector string

	switch o.WorkloadKind {
	case "deployment":
		o.WorkloadKind = Deployment
	case "statefulset":
		o.WorkloadKind = StatefulSet
	case "daemonset":
		o.WorkloadKind = DaemonSet
	default:
		o.WorkloadKind = ".*"
	}
	workloadSelector = fmt.Sprintf(`namespace="%s", workload=~"%s:(%s)"`, o.NamespaceName, o.WorkloadKind, o.ResourceFilter)

	if strings.Contains(metric, "deployment") {
		kindSelector = fmt.Sprintf(`namespace="%s", deployment!="", deployment=~"%s"`, o.NamespaceName, o.ResourceFilter)
	}
	if strings.Contains(metric, "statefulset") {
		kindSelector = fmt.Sprintf(`namespace="%s", statefulset!="", statefulset=~"%s"`, o.NamespaceName, o.ResourceFilter)
	}
	if strings.Contains(metric, "daemonset") {
		kindSelector = fmt.Sprintf(`namespace="%s", daemonset!="", daemonset=~"%s"`, o.NamespaceName, o.ResourceFilter)
	}

	return strings.NewReplacer("$1", workloadSelector, "$2", kindSelector).Replace(tmpl)
}

func makePodMetricExpr(tmpl string, o monitoring.QueryOptions) string {
	var podSelector, workloadSelector string

	// For monitoriong pods of the specific workload
	// GET /namespaces/{namespace}/workloads/{kind}/{workload}/pods
	if o.WorkloadName != "" {
		switch o.WorkloadKind {
		case "deployment":
			workloadSelector = fmt.Sprintf(`owner_kind="ReplicaSet", owner_name=~"^%s-[^-]{1,10}$"`, o.WorkloadName)
		case "statefulset":
			workloadSelector = fmt.Sprintf(`owner_kind="StatefulSet", owner_name="%s"`, o.WorkloadName)
		case "daemonset":
			workloadSelector = fmt.Sprintf(`owner_kind="DaemonSet", owner_name="%s"`, o.WorkloadName)
		}
	}

	// For monitoring pods in the specific namespace
	// GET /namespaces/{namespace}/workloads/{kind}/{workload}/pods or
	// GET /namespaces/{namespace}/pods/{pod} or
	// GET /namespaces/{namespace}/pods
	if o.NamespaceName != "" {
		if o.PodName != "" {
			podSelector = fmt.Sprintf(`pod="%s", namespace="%s"`, o.PodName, o.NamespaceName)
		} else {
			podSelector = fmt.Sprintf(`pod=~"%s", namespace="%s"`, o.ResourceFilter, o.NamespaceName)
		}
	} else {
		var namespaces, pods []string
		if o.NamespacedResourcesFilter != "" {
			for _, np := range strings.Split(o.NamespacedResourcesFilter, "|") {
				if nparr := strings.SplitN(np, "/", 2); len(nparr) > 1 {
					namespaces = append(namespaces, nparr[0])
					pods = append(pods, nparr[1])
				} else {
					pods = append(pods, np)
				}
			}
		}
		// For monitoring pods on the specific node
		// GET /nodes/{node}/pods/{pod}
		// GET /nodes/{node}/pods
		if o.NodeName != "" {
			if o.PodName != "" {
				if nparr := strings.SplitN(o.PodName, "/", 2); len(nparr) > 1 {
					podSelector = fmt.Sprintf(`namespace="%s",pod="%s", node="%s"`, nparr[0], nparr[1], o.NodeName)
				} else {
					podSelector = fmt.Sprintf(`pod="%s", node="%s"`, o.PodName, o.NodeName)
				}
			} else {
				var ps []string
				ps = append(ps, fmt.Sprintf(`node="%s"`, o.NodeName))
				if o.ResourceFilter != "" {
					ps = append(ps, fmt.Sprintf(`pod=~"%s"`, o.ResourceFilter))
				}

				if len(namespaces) > 0 {
					ps = append(ps, fmt.Sprintf(`namespace=~"%s"`, strings.Join(namespaces, "|")))
				}
				if len(pods) > 0 {
					ps = append(ps, fmt.Sprintf(`pod=~"%s"`, strings.Join(pods, "|")))
				}
				podSelector = strings.Join(ps, ",")
			}
		} else {
			// For monitoring pods in the whole cluster
			// Get /pods
			var ps []string
			if len(namespaces) > 0 {
				ps = append(ps, fmt.Sprintf(`namespace=~"%s"`, strings.Join(namespaces, "|")))
			}
			if len(pods) > 0 {
				ps = append(ps, fmt.Sprintf(`pod=~"%s"`, strings.Join(pods, "|")))
			}
			if len(ps) > 0 {
				podSelector = strings.Join(ps, ",")
			}
		}
	}

	return strings.NewReplacer("$1", workloadSelector, "$2", podSelector).Replace(tmpl)
}

func makeContainerMetricExpr(tmpl string, o monitoring.QueryOptions) string {
	var containerSelector string
	if o.ContainerName != "" {
		containerSelector = fmt.Sprintf(`pod="%s", namespace="%s", container="%s"`, o.PodName, o.NamespaceName, o.ContainerName)
	} else {
		containerSelector = fmt.Sprintf(`pod="%s", namespace="%s", container=~"%s"`, o.PodName, o.NamespaceName, o.ResourceFilter)
	}
	return strings.Replace(tmpl, "$1", containerSelector, -1)
}
