# quick-start

## Prerequisities

- A k8s CLuster(Kubernetes version > v1.18.x < v1.24.x) (could start with [kind](https://github.com/kubernetes-sigs/kind))
- Go (>1.17.x)
- kubectl (>1.19.x)
- Docker(18.x)

## start a k8s-cluster
if didnt have a k8s cluster, use `kind` to start

### install go and docker
 in a machine or vm,install `go` and `docker` and `kubectl`

### install kind

```shell
curl -Lo ./kind "https://kind.sigs.k8s.io/dl/v0.14.0/kind-$(uname)-amd64"
chmod +x ./kind
mv ./kind /usr/local/bin/kind
kind version
```

### config kind
make some cluser-config (`mykind.yml`) to start kind  :
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 30001
      hostPort: 40001
      listenAddress: "0.0.0.0"
    - containerPort: 30002
      hostPort: 40002
      listenAddress: "0.0.0.0"
    - containerPort: 30003
      hostPort: 40003
      listenAddress: "0.0.0.0"
    - containerPort: 30004
      hostPort: 40004
      listenAddress: "0.0.0.0"

```

this config will tranport localhost's 4000x to kind-cluster's 3000x, for visiting some cluster's service in local.you can edit with your own need.

### kind start cluster

```shell
# --config spectify  kind-config 
# --name spectify cluster-name (kind will add a kind- prefix)
# --image spectify the k8s version 
$ kind  create cluster --config mykind.yml --name mycluster  --image kindest/node:v1.19.9

Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.19.9) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-mycluster"
You can now use your cluster with:kubectl cluster-info --context kind-kind Have a nice day! 👋
```

### access kind-cluser

kind-cluster config will generate at localhost's `~/.kube/config`, if create multi cluster with kind, it will have multi contexts,so you should switch the context what you need to access

```shell
# list kind clusters
$ kind get clusters
mycluster

# get-contexts and switch, check where the current context is 
$ kubectl config get-contexts
CURRENT   NAME         CLUSTER      AUTHINFO     NAMESPACE
          kind-mycluster   kind-mycluster   kind-mycluster   
*         minikube     minikube     minikube     default

$ kubectl config set-context kind-mycluster
Context "kind-mycluster" modified.
```

now your can use ```kubectl``` access the cluster you create just as local luster

```shell
$ kubectl get ns
NAME                   STATUS   AGE
captain-system         Active   22h
default                Active   25h
kube-node-lease        Active   25h
kube-public            Active   25h
kube-system            Active   25h
kubernetes-dashboard   Active   23h
local-path-storage     Active   25h
```

## apply captain-server in cluster

### clone repo

```
git clone https://github.com/cuboss/captain.git
```

### check the ```deploy.yaml```

```wget https://github.com/cuboss/captain/blob/main/deploy/deploy.yaml```

### apply deploy

```shell
kubectl apply -f deploy.yaml
```

### check deploy and svc

```shell
# check po
$ kubectl get po -n captain-system
NAME                              READY   STATUS    RESTARTS   AGE
captain-server-7b6ccd677f-kg589   1/1     Running   0          127m

# check svc
$ kubectl get svc captain-server -n captain-system -o yaml        
apiVersion: v1
kind: Service
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"captain-server"},"name":"captain-server","namespace":"captain-system"},"spec":{"ports":[{"name":"http","port":9090,"targetPort":9090}],"selector":{"app":"captain-server"}}}
  creationTimestamp: "2022-07-27T09:47:39Z"
  labels:
    app: captain-server
  name: captain-server
  namespace: captain-system
  resourceVersion: "16259"
  uid: 179b7c41-a00a-4b8f-96c6-d482c10da091
spec:
  clusterIP: 10.96.72.139
  clusterIPs:
  - 10.96.72.139
  externalTrafficPolicy: Cluster
  internalTrafficPolicy: Cluster
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - name: http
    port: 9090
    protocol: TCP
    targetPort: 9090
  selector:
    app: captain-server
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}

```

if you want to visit captain-server out of cluster ,you cloud edit the svc with NodePort type

```diff
# captain-server svc
spec:
  - name: http
+   nodePort: 30002
    port: 9090
    protocol: TCP
    targetPort: 9090
  selector:
    app: captain-server
-  type: ClusterIP
+  type: NodePort

```

now you can access localhost's 40002 to visit kind-cluster's 30002 (we configed at creating kind-cluster before).

so now you can just curl get the servet-api 'sresult:

```shell
$curl curl localhost:40002/api/v1/namespaces
{
  "kind": "NamespaceList",
  "apiVersion": "v1",
  "metadata": {
    "resourceVersion": "25950"
  },
  "items": [
    {
      "metadata": {
        "name": "captain-system",
...
      }
    }]
}

# if you install jq , then you can parse the json to choose result
$ curl 127.0.0.1:40002/api/v1/namespaces |  jq -rM '.items[].metadata.name'
captain-system
default
kube-node-lease
kube-public
kube-system
kubernetes-dashboard
local-path-storage

```
