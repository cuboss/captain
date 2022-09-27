# 多集群
## 接口
GET /capis/cluster.captain.io/clusters\
GET /capis/cluster.captain.io/clusters/{name}\
POST /capis/cluster.captain.io/clusters\
DELETE /capis/cluster.captain.io/clusters/{name}\
PUT /capis/cluster.captain.io/clusters/{name}

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