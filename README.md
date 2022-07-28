# captain

This project aims to building a platform which has rich function of managing k8s clusters in everywhere network would be reach.
## Overview

[![codecov](https://codecov.io/gh/cuboss/captain/branch/main/graph/badge.svg)](https://codecov.io/gh/cuboss/captain)
[![Release](https://img.shields.io/github/v/release/cuboss/captain)](https://img.shields.io/github/v/release/cuboss/captain)

|                                             **Stargazers Over Time**                                              | **Contributors Over Time**                                                                                                                                                                                                                       |
|:-----------------------------------------------------------------------------------------------------------------:|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|      [![Stargazers over time](https://starchart.cc/cuboss/captain.svg)](https://starchart.cc/cuboss/captain)      | [![Contributor over time](https://contributor-graph-api.apiseven.com/contributors-svg?chart=contributorOverTime&repo=cuboss/captain)](https://contributor-graph-api.apiseven.com/contributors-svg?chart=contributorOverTime&repo=cuboss/captain) |

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

### 编译镜像

```shell
make image
```

### 请求

```shell
# 请求namespace 等api/v1的资源
curl http://127.0.0.1:9090/api/v1/namespaces 
# 请求deployment 等apis/apps/v1的资源
curl http://127.0.0.1:9090/apis/apps/v1/deployments
```
