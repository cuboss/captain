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

level:
- cluster
- node
- workload
- pod
## 参数
metrics_filter: 查询参数，value是一个正则表达式，指明要查询的一个或多个指标名称。
```
eg:
    metrics_filter=cluster_cpu_usage|cluster_cpu_total|cluster_memory_usage_wo_cache|cluster_memory_total|cluster_disk_size_usage|cluster_disk_size_capacity|cluster_pod_running_count|cluster_pod_quota$
```
当前支持的指标名称：
- cluster_cpu_usage
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