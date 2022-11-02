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