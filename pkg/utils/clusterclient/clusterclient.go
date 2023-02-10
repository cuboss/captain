package clusterclient

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	clusterv1alpha1 "captain/apis/cluster/v1alpha1"
	clusterinformer "captain/pkg/client/informers/externalversions/cluster/v1alpha1"
	"captain/pkg/simple/client/multicluster"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	ClusterNotExistsFormat = "cluster %s not exists"
)

type innerCluster struct {
	KubernetesURL *url.URL
	CaptainURL    *url.URL
	Transport     http.RoundTripper
}

type ClusterClients interface {
	IsHostCluster(cluster *clusterv1alpha1.Cluster) bool
	IsClusterReady(cluster *clusterv1alpha1.Cluster) bool
	GetClusterKubeconfig(string) (string, error)
	Get(region, cluster string) (*clusterv1alpha1.Cluster, error)
	GetByClusterName(clustername string) (*clusterv1alpha1.Cluster, error)
	GetInnerCluster(string) *innerCluster
	GetClientSet(string, string) (*kubernetes.Clientset, error)
}

type clusterClients struct {
	sync.RWMutex
	clusterMap        map[string]*clusterv1alpha1.Cluster
	clusterKubeconfig map[string]string

	// build a in memory cluster cache to speed things up
	innerClusters map[string]*innerCluster

	options *multicluster.Options
}

func (c *clusterClients) IsHostCluster(cluster *clusterv1alpha1.Cluster) bool {
	if _, ok := cluster.Labels[clusterv1alpha1.HostCluster]; ok {
		return true
	}
	return false
}

func (c *clusterClients) IsClusterReady(cluster *clusterv1alpha1.Cluster) bool {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == clusterv1alpha1.ClusterReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (c *clusterClients) GetClusterKubeconfig(clusterName string) (string, error) {
	c.RLock()
	defer c.RUnlock()
	if c, exists := c.clusterKubeconfig[clusterName]; exists {
		return c, nil
	} else {
		return "", fmt.Errorf(ClusterNotExistsFormat, clusterName)
	}
}

func (c *clusterClients) Get(regionName, clusterName string) (*clusterv1alpha1.Cluster, error) {
	if regionName == c.options.HostRegionName {
		regionName = ""
	}

	if len(regionName) > 0 {
		clusterName = fmt.Sprintf("%s-%s", regionName, clusterName)
	}
	return c.GetByClusterName(clusterName)
}

func (c *clusterClients) GetByClusterName(clusterName string) (*clusterv1alpha1.Cluster, error) {
	c.RLock()
	defer c.RUnlock()
	if cluster, exists := c.clusterMap[clusterName]; exists {
		return cluster, nil
	} else {
		return nil, fmt.Errorf(ClusterNotExistsFormat, clusterName)
	}
}

func (c *clusterClients) GetInnerCluster(name string) *innerCluster {
	c.RLock()
	defer c.RUnlock()
	if cluster, ok := c.innerClusters[name]; ok {
		return cluster
	}
	return nil
}

func (c *clusterClients) GetClientSet(regionName, clusterName string) (*kubernetes.Clientset, error) {
	// TODO cache
	cluster, err := c.Get(regionName, clusterName)
	if err != nil {
		return nil, err
	}
	r, err := clientcmd.RESTConfigFromKubeConfig([]byte(cluster.Spec.Connection.KubeConfig))
	if err != nil {
		return nil, fmt.Errorf("get cluster kubeconfig restconfig err: %v", err)
	}
	return kubernetes.NewForConfig(r)
}

var c *clusterClients
var lock sync.Mutex

func NewClusterClients(clusterInformer clusterinformer.ClusterInformer, options *multicluster.Options) ClusterClients {

	if c == nil {
		lock.Lock()
		defer lock.Unlock()

		if c != nil {
			return c
		}

		c = &clusterClients{
			clusterMap:        map[string]*clusterv1alpha1.Cluster{},
			clusterKubeconfig: map[string]string{},
			innerClusters:     make(map[string]*innerCluster),
			options:           options,
		}

		clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.addCluster(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				c.removeCluster(oldObj)
				c.addCluster(newObj)
			},
			DeleteFunc: func(obj interface{}) {
				c.removeCluster(obj)
			},
		})
	}

	return c
}

func (c *clusterClients) removeCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)
	klog.V(4).Infof("remove cluster %s", cluster.Name)
	c.Lock()
	if _, ok := c.clusterMap[cluster.Name]; ok {
		delete(c.clusterMap, cluster.Name)
		delete(c.innerClusters, cluster.Name)
		delete(c.clusterKubeconfig, cluster.Name)
	}
	c.Unlock()
}

func (c *clusterClients) addCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)
	klog.V(4).Infof("add new cluster %s", cluster.Name)
	_, err := url.Parse(cluster.Spec.Connection.KubernetesAPIEndpoint)
	if err != nil {
		klog.Errorf("Parse kubernetes apiserver endpoint %s failed, %v", cluster.Spec.Connection.KubernetesAPIEndpoint, err)
		return
	}

	innerCluster := newInnerCluster(cluster)
	c.Lock()
	c.clusterMap[cluster.Name] = cluster
	c.clusterKubeconfig[cluster.Name] = string(cluster.Spec.Connection.KubeConfig)
	c.innerClusters[cluster.Name] = innerCluster
	c.Unlock()
}

func newInnerCluster(cluster *clusterv1alpha1.Cluster) *innerCluster {
	kubernetesEndpoint, err := url.Parse(cluster.Spec.Connection.KubernetesAPIEndpoint)
	if err != nil {
		klog.Errorf("Parse kubernetes apiserver endpoint %s failed, %v", cluster.Spec.Connection.KubernetesAPIEndpoint, err)
		return nil
	}

	captainEndpoint, err := url.Parse(cluster.Spec.Connection.CaptainAPIEndpoint)
	if err != nil {
		klog.Errorf("Parse captain apiserver endpoint %s failed, %v", cluster.Spec.Connection.CaptainAPIEndpoint, err)
		return nil
	}

	// prepare for
	clientConfig, err := clientcmd.NewClientConfigFromBytes(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		klog.Errorf("Unable to create client config from kubeconfig bytes, %#v", err)
		return nil
	}

	clusterConfig, err := clientConfig.ClientConfig()
	if err != nil {
		klog.Errorf("Failed to get client config, %#v", err)
		return nil
	}

	if len(cluster.Spec.Connection.KubernetesAPIEndpoint) == 0 {
		kubernetesEndpoint, _ = url.Parse(clusterConfig.Host)
	}

	transport, err := rest.TransportFor(clusterConfig)
	if err != nil {
		klog.Errorf("Create transport failed, %v", err)
		return nil
	}

	return &innerCluster{
		KubernetesURL: kubernetesEndpoint,
		CaptainURL:    captainEndpoint,
		Transport:     transport,
	}
}
