# 多集群
需要在host集群上上先 apply `clusters`的crd: [cluster.captain.io_clusters.yaml](../deploy/crd/cluster/cluster.captain.io_clusters.yaml)
## 接口
### 获取cluster列表
GET `/capis/cluster.captain.io/v1alpha1/clusters` \
Request: 
| 参数      | 参数类型 | 数据类型 | 说明               |
| --------- | -------- | -------- | ------------------ |
| page      | query    | int      | 页码               |
| pageSize  | query    | int      | 页大小             |
| sortBy    | query    | string   | 按照哪个字段排序   |
| ascending | query    | boolean  | 排序参数 默认false |

Response:
```json
{
 "items": [
   {}
 ],
 "totalItems": 2,
 "pageSize": 10,
 "totalPages": 1,
 "currentPage": 1
}
```
其中`items`元素为cluster数据，cluster数据结构参考文档末尾。
### 获取指定cluster
GET /capis/cluster.captain.io/v1alpha1/clusters/{name}

Request: 
| 参数      | 参数类型 | 数据类型 | 说明               |
| --------- | -------- | -------- | ------------------ |
| name      | path    | string      | cluster名称      |

Response: `Cluster{}`

### 创建（纳管）cluster
POST /capis/cluster.captain.io/v1alpha1/clusters

Request: 
| 参数      | 参数类型 | 数据类型 | 说明               |
| --------- | -------- | -------- | ------------------ |
| body      | body    | Cluster      | 集群信息      |

Response: `Cluster{}`

### 移除（取消纳管）cluster
DELETE /capis/cluster.captain.io/v1alpha1/clusters/{name}

Request: 
| 参数      | 参数类型 | 数据类型 | 说明               |
| --------- | -------- | -------- | ------------------ |
| name      | path    | string      | cluster名称      |

### 修改cluster
PUT /capis/cluster.captain.io/v1alpha1/clusters/{name}

Request: 
| 参数      | 参数类型 | 数据类型 | 说明               |
| --------- | -------- | -------- | ------------------ |
| name      | path    | string      | cluster名称      |

Response: `Cluster{}`

## 多集群代理接口
/regions/{region}/cluster/{name}/...\
eg. 
```bash
curl http://127.0.0.1:9090/regions/wx-tst/clusters/cke-tst/api/v1/namespaces
```

## 注意
创建Cluster时：
+ cluster.Name添加前缀 {region}-， 如cluster1->xxtst-cluster1。
+ 添加region的label，cluster.captain.io/region: {region}，如cluster.captain.io/region: xxtst

*Why?*

不同region下可能存在同名的cluster，为了避免冲突，在存在region的情况下，使用region作为前缀。\
再此情况下为了能获取到正确的集群名称，同时需要在label中标明region字段，前端或代码中获取cluster名称的逻辑为当存在`cluster.captain.io/region` Label，且value不为空的情况下，集群名称为cluster.Name再去除{region}-前缀。label不存在或value为空直接返回cluster.Name

多集群功能应该允许region为空的情况，比如私有云单云池场景。所以也应该支持这种场景下对应的api路径，例如/clusters/{cluster}/...



# cluster结构
```json
{
 "kind": "Cluster",
 "apiVersion": "cluster.captain.io/v1alpha1",
 "metadata": {
  "name": "wx-tst-cke-tst",
  "labels": {
   "cluster.captain.io/region": "wx-tst"
  },
 "spec": {
  "enable": true,
  "provider": "CKE",
  "connection": {
   "type": "direct",
   "kubernetesAPIEndpoint": "http://xx.xx.xx.xx:6443",
   "kubeconfig": "xxxx(base64 []byte)"
  }
 },
 "status": {}
}
```


