# 监控模块实现思考
## ks监控模块参考
kubersphere监控模块主要是对prometheus进行请求代理，多集群的监控指标请求则是由master集群的ks-apiserver转发至子集群的ks-apiserver，再由子集群的ks-apiserver到指定的prometheus endpoint中查询。

## 我们该如何做？
## 方案一
```yml
prometheus:
  endpoints: xxxx
  auth:
    basic:
      username: xxxx
      password: xxxx(base64)
```
把子集群的prometheus的服务地址和端口配置在cluster资源中，由管理集群的captain直接访问该地址端口

优点：不存在apiserver和prometheus的Authorization冲突的问题

缺点：需要在子集群配置prometheus服务为集群外可访问
## 方案二
```yml
prometheus:
  service:
    protocol: http #default http
    namespace: monitoring
    name: prometheus-k8s
    port: 9090
  auth:
    basic:
      username: xxxx
      password: xxxx(base64)
```
在cluster资源中配置子集群的prometheus service信息，captain通过apiserver的service proxy接口访问集群内的prometheus service

优点：不需要在子集群配置prometheus服务为集群外可访问

缺点：存在apiserver和prometheus的Authorization冲突的问题
## 方案三
类似ks，在子集群安装captain-server或者监控/日志代理，由管理集群的captain把请求转发至子集群的代理组件

优点：不需要在子集群配置prometheus服务为集群外可访问，不存在apiserver和prometheus的Authorization冲突的问题

缺点：需要在子集群安装代理组件

基于日志采集肯定是要放在子集群代理组件进行，所有不妨让代理组件也兼具指标采集功能。倾向于选择方案三。

需要实现自动安装captain组件。

## 当前需要的监控指标
1. cluster
   - 资源使用情况
       - CPU
       - 内存
       - 磁盘组
   - 节点资源使用量TOPK
       - CPU
       - 内存
       - 磁盘组
       - pod数量
2. node
   - CPU用量
   - 内存用量
   - 容器组数量
3. workload
   - CPU
   - 内存
4. Pod
   - CPU
   - 内存

# 接口设计
GET /capis/monitoring.captain.io/v1alpha1/{level}

GET /capis/monitoring.captain.io/v1alpha1/{level}/{resourcename}

level:
- cluster
- nodes
- workload
- pod
## 参数
| 参数名 | 描述 |
| --- | --- |
| 指标/资源查询 |
| metrics_filter | string<br>用于指定查询的指标名称，是一个正则表达式。比如同时查询节点cpu和磁盘用量:  `node_cpu_usage\|node_disk_size_usage$`. 支持的指标可以参考 metrices.md . |
| resources_filter | string<br>用于指定查询的资源名称，是一个正则表达式。比如查询节点i-caojnter 和 i-cmu82og: `i-caojnter\|i-cmu82ogj$`. |
| 范围查询 |  |
| start | string<br>查询指定时间范围内指标时，start和end表示开始和结束的时间戳， Unix时间格式，如：1559347200. |
| end | string<br>查询指定时间范围内指标时，start和end表示开始和结束的时间戳， Unix时间格式，如：1559347200. |
| step | string<br>Default: "10m"<br>在start和end范围中，以step为时间间隔检索数据。 格式为 [0-9]+[smhdwy]. 默认为10m。 |
| 指定时间 |  |
| time | string<br>Unix时间格式的时间戳。检索单个时间点的度量数据。如果为空，默认为当前时间。time和start、end、step的组合是互斥的。 |
| 分页排序 | |
| sort_metric | string<br>根据指定的指标进行排序。不适用于提供了start和end参数的情况 |
| sort_type | string<br>Default: "desc."<br>排序类型，升序/降序. One of asc, desc. |
| page | integer<br>页码。这个字段对每个指标的结果数据进行分页，然后返回一个特定的页面。例如，将page设置为2将返回第二个页面。它只适用于排序的度量数据。 |
| limit | integer<br>每页返回的记录条数，默认值5 |

注： metrics_filter: 查询参数，value是一个正则表达式，指明要查询的一个或多个指标名称。
```
eg:
    metrics_filter=cluster_cpu_usage|cluster_cpu_total|cluster_memory_usage_wo_cache|cluster_memory_total|cluster_disk_size_usage|cluster_disk_size_capacity|cluster_pod_running_count|cluster_pod_quota$
```
当前支持的指标名称：
    参见metrics.md

返回内容：

精确时间点查询：
```json
{
    "results": [
        {
            "metric_name": "node_pod_count",
            "data": {
                "resultType": "vector",
                "result": [
                    {
                        "metric": {
                            "__name__": "node:pod_count:sum",
                            "node": "192.168.0.1"
                        },
                        "value": [
                            1666856591.888,
                            "28"
                        ],
                        "min_value": "",
                        "max_value": "",
                        "avg_value": "",
                        "sum_value": "",
                        "fee": "",
                        "resource_unit": "",
                        "currency_unit": ""
                    }
                ]
            }
        },
    ]
}
```

范围查询：
```json
{
    "results": [
        {
            "metric_name": "cluster_cpu_total",
            "data": {
                "resultType": "matrix",
                "result": [
                    {
                        "values": [
                            [
                                1666856634,
                                "56"
                            ],
                            [
                                1666856934,
                                "56"
                            ],
                            [
                                1666857234,
                                "56"
                            ]
                        ],
                        "min_value": "",
                        "max_value": "",
                        "avg_value": "",
                        "sum_value": "",
                        "fee": "",
                        "resource_unit": "",
                        "currency_unit": ""
                    }
                ]
            }
        }
    ]
}
```

GetNamedMetricsOverTime 和 GetNamedMetrics 差别？\
GetNamedMetrics 返回Vector，获取的是瞬时的指标数据。\
GetNamedMetricsOverTime 返回Matrix，获取的是一段时间内的多个指标数据。\
Vector和Matrix代表的都是名称相同但leabel不同的一组指标数据。
```yaml
Vector:
- metric:
    nodename: xx
	podname: xxx
  value:
    26.04
  timestamp:
    20221021112809

Matrix:
- metric:
  values:
  - Timestamp: 20221021112809
    Value: 26.04
```
