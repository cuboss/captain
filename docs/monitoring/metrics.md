| 指标名称  | label | value单位 | 描述 | promql |
|  ---- | ---- | ---- | ---- | ---- |
| cluster |
| cluster_cpu_utilisation | 无 | 百分比 | 集群CPU使用率 | :node_cpu_utilisation:avg1m |
| cluster_cpu_usage | 无 | Core | 集群CPU用量 | round(:node_cpu_utilisation:avg1m * sum(node:node_num_cpu:sum), 0.001) |
| cluster_cpu_total | 无 | Core | 集群CPU总数 | sum(node:node_num_cpu:sum) |
| cluster_memory_utilisation | 无 | 百分比 | 集群内存使用率 | :node_memory_utilisation: |
| cluster_memory_available | 无 | Byte | 集群可用内存 | sum(node:node_memory_bytes_available:sum) |
| cluster_memory_total | 无 | Byte | 集群内存总量 | sum(node:node_memory_bytes_total:sum) |
| cluster_disk_size_usage | 无 | Byte | 集群磁盘使用量 | sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"} - node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance)) |
| cluster_disk_size_utilisation | 无 | 百分比 | 集群磁盘使用率 | cluster:disk_utilization:ratio |
| cluster_disk_size_capacity | 无 | Byte | 集群磁盘总容量 | sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance)) |
| cluster_disk_size_available | 无 | Byte | 集群磁盘可用大小 | sum(max(node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance)) |
| node |
| node_cpu_utilisation | node| 百分比 | 节点 CPU 使用率 | node:node_cpu_utilisation:avg1m{$1} |
| node_cpu_usage | node| Core | 节点 CPU 用量 | round(node:node_cpu_utilisation:avg1m{$1} * node:node_num_cpu:sum{$1}, 0.001) |
| node_cpu_total | node | Core | 节点 CPU 总量 | node:node_num_cpu:sum{$1} |
| node_memory_utilisation | node| 百分比 | 节点内存使用率 | node:node_memory_utilisation:{$1} |
| node_memory_available | node| Byte | 节点可用内存 | node:node_memory_bytes_available:sum{$1} |
| node_memory_total | node| Byte | 节点内存总量 | node:node_memory_bytes_total:sum{$1} |
| node_memory_usage_wo_cache | node| Byte | 节点内存使用量 | node:node_memory_bytes_total:sum{$1} - node:node_memory_bytes_available:sum{$1} |
| node_pod_count | node| 个 | 节点调度完成 Pod 数量 | node:pod_count:sum{$1} |
| node_pod_quota | node | 个 | 节点 Pod 最大容纳量| max(kube_node_status_capacity{resource="pods",$1}) by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0) |
| workload |
| workload_cpu_usage |  | Core | 工作负载CPU 用量 | round(namespace:workload_cpu_usage:sum{$1}, 0.001) |
| workload_memory_usage |  | Byte | 工作负载内存使用量（包含缓存） | namespace:workload_memory_usage:sum{$1} |
| workload_memory_usage_wo_cache |  | Byte | 工作负载内存使用量 | namespace:workload_memory_usage_wo_cache:sum{$1} |
| workload_net_bytes_transmitted |  | Byte/s | 工作负载网络数据发送速率 | namespace:workload_net_bytes_transmitted:sum_irate{$1} |
| workload_net_bytes_received |  | Byte/s | 工作负载网络数据接受速率 | namespace:workload_net_bytes_received:sum_irate{$1} |
| workload_deployment_replica |  |  | Deployment 期望副本数 | label_join(sum (label_join(label_replace(kube_deployment_spec_replicas{$2}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload") |
| workload_deployment_replica_available |  |  | Deployment 可用副本数 | label_join(sum (label_join(label_replace(kube_deployment_status_replicas_available{$2}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload") |
| workload_statefulset_replica |  |  | StatefulSet 期望副本数 | label_join(sum (label_join(label_replace(kube_statefulset_replicas{$2}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload") |
| workload_statefulset_replica_available |  |  | StatefulSet 可用副本数 | label_join(sum (label_join(label_replace(kube_statefulset_status_replicas_current{$2}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload") |
| workload_daemonset_replica |  |  | DaemonSet 期望副本数 | label_join(sum (label_join(label_replace(kube_daemonset_status_desired_number_scheduled{$2}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload") |
| workload_daemonset_replica_available |  |  | DaemonSet 可用副本数 | label_join(sum (label_join(label_replace(kube_daemonset_status_number_available{$2}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload") |
| workload_deployment_unavailable_replicas_ratio |  |  | Deployment 不可用副本数比例 | namespace:deployment_unavailable_replicas:ratio{$1} |
| workload_daemonset_unavailable_replicas_ratio |  |  | DaemonSet 不可用副本数比例 | namespace:daemonset_unavailable_replicas:ratio{$1} |
| workload_statefulset_unavailable_replicas_ratio |  |  | StatefulSet 不可用副本数比例 | namespace:statefulset_unavailable_replicas:ratio{$1} |
| pod |
|pod_cpu_usage|  | Core | 容器组 CPU 用量 | round(sum by (namespace, pod) (irate(container_cpu_usage_seconds_total{job="kubelet", pod!="", image!=""}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}, 0.001) |
|pod_memory_usage|  | Byte | 容器组内存使用量（包含缓存） | sum by (namespace, pod) (container_memory_usage_bytes{job="kubelet", pod!="", image!=""}) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2} |
|pod_memory_usage_wo_cache|  | Byte | 容器组内存使用量 | sum by (namespace, pod) (container_memory_working_set_bytes{job="kubelet", pod!="", image!=""}) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2} |
|pod_net_bytes_transmitted|  | Byte/s | 容器组网络数据发送速率 | sum by (namespace, pod) (irate(container_network_transmit_bytes_total{pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2} |
|pod_net_bytes_received|  | Byte/s | 容器组网络数据接受速率 | sum by (namespace, pod) (irate(container_network_receive_bytes_total{pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2} |
| container |
| container_cpu_usage |  |  | 容器 CPU 用量 | round(sum by (namespace, pod, container) (irate(container_cpu_usage_seconds_total{job="kubelet", container!="POD", container!="", image!="", $1}[5m])), 0.001) |
| container_memory_usage |  |  | 容器内存使用量（包含缓存） | sum by (namespace, pod, container) (container_memory_usage_bytes{job="kubelet", container!="POD", container!="", image!="", $1}) |
| container_memory_usage_wo_cache |  |  | 容器内存使用量 | sum by (namespace, pod, container) (container_memory_working_set_bytes{job="kubelet", container!="POD", container!="", image!="", $1}) |
