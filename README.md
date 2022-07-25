# captain
This project aims to buillding a platform which has rich function of managing k8s clusters in everywhere network would be reach. 

## quick-start

### 编译
进入主目录
```shell
make captain-server
```

### 运行captain-server

```
./bin/captain-server
```
会连接本地集群 或者指定kubeconfig

### 请求

```shell
# 请求namespace 等api/v1的资源
curl http://127.0.0.1:9090/api/v1/namespaces 
# 请求deployment 等apis/apps/v1的资源
curl http://127.0.0.1:9090/apis/apps/v1/deployments
```
