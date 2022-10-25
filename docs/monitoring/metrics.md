| 指标名称  | label | value单位 | 描述 | promql |
|  ---- | ---- | ---- | ---- | ---- |
| cluster_cpu_utilisation | 无 | 百分比 | 集群CPU使用率 | :node_cpu_utilisation:avg1m |
| cluster_cpu_usage | 无 | Core | 集群CPU用量 | round(:node_cpu_utilisation:avg1m * sum(node:node_num_cpu:sum), 0.001) |
| cluster_cpu_total | 无 | Core | 集群CPU总数 | sum(node:node_num_cpu:sum) |
| cluster_memory_utilisation | 无 | 百分比 | 集群内存使用率 | :node_memory_utilisation: |
| cluster_memory_available | 无 | Byte | 集群可用内存 | sum(node:node_memory_bytes_available:sum) |
| cluster_memory_total | 无 | Byte | 集群内存总量 | sum(node:node_memory_bytes_total:sum) |
