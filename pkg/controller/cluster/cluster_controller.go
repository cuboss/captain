package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	clusterv1alpha1 "captain/apis/cluster/v1alpha1"
	"captain/pkg/client/clientset/versioned/scheme"
	clusterclient "captain/pkg/client/clientset/versioned/typed/cluster/v1alpha1"
	clusterinformer "captain/pkg/client/informers/externalversions/cluster/v1alpha1"
	clusterlister "captain/pkg/client/listers/cluster/v1alpha1"
	"captain/pkg/version"
)

// Cluster controller only runs under multicluster mode. Cluster controller is following below steps,
//   1. Wait for cluster agent is ready if connection type is proxy
//   2. Join cluster into federation control plane if kubeconfig is ready.
//   3. Pull cluster version and configz, set result to cluster status
// Also put all clusters back into queue every 5 * time.Minute to sync cluster status, this is needed
// in case there aren't any cluster changes made.
// Also check if all of the clusters are ready by the spec.connection.kubeconfig every resync period

const (
	// maxRetries is the number of times a service will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15

	kubefedNamespace  = "kube-federation-system"
	openpitrixRuntime = "openpitrix.io/runtime"
	captainManaged    = "captain.io/managed"

	// proxy format
	proxyFormat = "%s/api/v1/namespaces/captain-system/services/:captain-apiserver:80/proxy/%s"

	// mulitcluster configuration name
	configzMultiCluster = "multicluster"

	// probe cluster timeout
	probeClusterTimeout = 3 * time.Second
)

// Cluster template for reconcile host cluster if there is none.
var hostCluster = &clusterv1alpha1.Cluster{
	ObjectMeta: metav1.ObjectMeta{
		Name: "host",
		Annotations: map[string]string{
			"captainManaged.io/description": "The description was created by Captain automatically. " +
				"It is recommended that you use the Host Cluster to manage clusters only " +
				"and deploy workloads on Member Clusters.",
		},
		Labels: map[string]string{
			clusterv1alpha1.HostCluster: "",
			captainManaged:              "true",
		},
	},
	Spec: clusterv1alpha1.ClusterSpec{
		// JoinFederation: true,
		Enable:   true,
		Provider: "captain",
		Connection: clusterv1alpha1.Connection{
			Type: clusterv1alpha1.ConnectionTypeDirect,
		},
	},
}

// ClusterData stores cluster client
type clusterData struct {

	// cached rest.Config
	config *rest.Config

	// cached kubernetes client, rebuild once cluster changed
	client kubernetes.Interface

	// cached kubeconfig
	cachedKubeconfig []byte

	// cached transport, used to proxy captain version request
	transport http.RoundTripper
}

type clusterController struct {
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	// build this only for host cluster
	client     kubernetes.Interface
	hostConfig *rest.Config

	clusterClient clusterclient.ClusterInterface

	clusterLister    clusterlister.ClusterLister
	clusterHasSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration

	mu sync.RWMutex

	clusterMap map[string]*clusterData

	resyncPeriod time.Duration

	hostClusterNmae string
}

func NewClusterController(
	client kubernetes.Interface,
	config *rest.Config,
	clusterInformer clusterinformer.ClusterInformer,
	clusterClient clusterclient.ClusterInterface,
	resyncPeriod time.Duration,
	hostClusterName string,
) *clusterController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cluster-controller"})

	c := &clusterController{
		eventBroadcaster: broadcaster,
		eventRecorder:    recorder,
		client:           client,
		hostConfig:       config,
		clusterClient:    clusterClient,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster"),
		workerLoopPeriod: time.Second,
		clusterMap:       make(map[string]*clusterData),
		resyncPeriod:     resyncPeriod,
		hostClusterNmae:  hostClusterName,
	}
	c.clusterLister = clusterInformer.Lister()
	c.clusterHasSynced = clusterInformer.Informer().HasSynced

	clusterInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: c.addCluster,
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.addCluster(newObj)
		},
		DeleteFunc: c.addCluster,
	}, resyncPeriod)

	return c
}

func (c *clusterController) Start(ctx context.Context) error {
	fmt.Println("Started")
	return c.Run(3, ctx.Done())
}

func (c *clusterController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.V(0).Info("starting cluster controller")
	defer klog.Info("shutting down cluster controller")

	if !cache.WaitForCacheSync(stopCh, c.clusterHasSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, c.workerLoopPeriod, stopCh)
	}

	// refresh cluster configz every resync period
	go wait.Until(func() {
		if err := c.reconcileHostCluster(); err != nil {
			klog.Errorf("Error create host cluster, error %v", err)
		}

		if err := c.probeClusters(); err != nil {
			klog.Errorf("failed to reconcile cluster ready status, err: %v", err)
		}

	}, c.resyncPeriod, stopCh)

	<-stopCh
	return nil
}

func (c *clusterController) worker() {
	for c.processNextItem() {
	}
}

func (c *clusterController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.syncCluster(key.(string))
	c.handleErr(err, key)
	return true
}

func buildClusterData(kubeconfig []byte) (*clusterData, error) {
	// prepare for
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		klog.Errorf("Unable to create client config from kubeconfig bytes, %#v", err)
		return nil, err
	}

	clusterConfig, err := clientConfig.ClientConfig()
	if err != nil {
		klog.Errorf("Failed to get client config, %#v", err)
		return nil, err
	}

	transport, err := rest.TransportFor(clusterConfig)
	if err != nil {
		klog.Errorf("Failed to create transport, %#v", err)
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		klog.Errorf("Failed to create ClientSet from config, %#v", err)
		return nil, err
	}

	return &clusterData{
		cachedKubeconfig: kubeconfig,
		config:           clusterConfig,
		client:           clientSet,
		transport:        transport,
	}, nil
}

// reconcileHostCluster will create a host cluster if there are no clusters labeled 'cluster-role.captain.io/host'
func (c *clusterController) reconcileHostCluster() error {
	clusters, err := c.clusterLister.List(labels.SelectorFromSet(labels.Set{clusterv1alpha1.HostCluster: ""}))
	if err != nil {
		return err
	}

	hostKubeConfig, err := buildKubeconfigFromRestConfig(c.hostConfig)
	if err != nil {
		return err
	}

	// no host cluster, create one
	if len(clusters) == 0 {
		hostCluster.Spec.Connection.KubeConfig = hostKubeConfig
		hostCluster.Name = c.hostClusterNmae
		_, err = c.clusterClient.Create(context.TODO(), hostCluster, metav1.CreateOptions{})
		return err
	} else if len(clusters) > 1 {
		return fmt.Errorf("there MUST not be more than one host clusters, while there are %d", len(clusters))
	}

	// only deal with cluster managed by captain
	cluster := clusters[0]
	managedByCaptain, ok := cluster.Labels[captainManaged]
	if !ok || managedByCaptain != "true" {
		return nil
	}

	// no kubeconfig, not likely to happen
	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		cluster.Spec.Connection.KubeConfig = hostKubeConfig
	} else {
		// if kubeconfig are the same, then there is nothing to do
		if bytes.Equal(cluster.Spec.Connection.KubeConfig, hostKubeConfig) {
			return nil
		}
	}

	// update host cluster config
	_, err = c.clusterClient.Update(context.TODO(), cluster, metav1.UpdateOptions{})
	return err
}

func (c *clusterController) probeClusters() error {
	clusters, err := c.clusterLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		// if the cluster is not federated, we skip it and consider it not ready.
		if !isConditionTrue(cluster, clusterv1alpha1.ClusterFederated) {
			continue
		}

		if len(cluster.Spec.Connection.KubeConfig) == 0 {
			continue
		}

		clientConfig, err := clientcmd.NewClientConfigFromBytes(cluster.Spec.Connection.KubeConfig)
		if err != nil {
			klog.Error(err)
			continue
		}

		config, err := clientConfig.ClientConfig()
		if err != nil {
			klog.Error(err)
			continue
		}
		config.Timeout = probeClusterTimeout

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			klog.Error(err)
			continue
		}

		var con clusterv1alpha1.ClusterCondition
		_, err = clientSet.Discovery().ServerVersion()
		if err == nil {
			con = clusterv1alpha1.ClusterCondition{
				Type:               clusterv1alpha1.ClusterReady,
				Status:             v1.ConditionTrue,
				LastUpdateTime:     metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             string(clusterv1alpha1.ClusterReady),
				Message:            "Cluster is available now",
			}
		} else {
			con = clusterv1alpha1.ClusterCondition{
				Type:               clusterv1alpha1.ClusterReady,
				Status:             v1.ConditionFalse,
				LastUpdateTime:     metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             "failed to connect get kubernetes version",
				Message:            "Cluster is not available now",
			}
		}

		c.updateClusterCondition(cluster, con)
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			ct, err := c.clusterClient.Get(context.TODO(), cluster.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			ct.Status.Conditions = cluster.Status.Conditions
			ct, err = c.clusterClient.Update(context.TODO(), ct, metav1.UpdateOptions{})
			return err
		})
		if err != nil {
			klog.Errorf("failed to update cluster %s status, err: %v", cluster.Name, err)
		} else {
			klog.V(4).Infof("successfully updated cluster %s to status %v", cluster.Name, con)
		}

	}

	return nil
}

func (c *clusterController) syncCluster(key string) error {
	klog.V(5).Infof("starting to sync cluster %s", key)
	startTime := time.Now()
	fmt.Println(key)
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("not a valid controller key %s, %#v", key, err)
		return err
	}

	defer func() {
		klog.V(4).Infof("Finished syncing cluster %s in %s", name, time.Since(startTime))
	}()

	cluster, err := c.clusterLister.Get(name)

	if err != nil {
		// cluster not found, possibly been deleted
		// need to do the cleanup
		if errors.IsNotFound(err) {
			return nil
		}

		klog.Errorf("Failed to get cluster with name %s, %#v", name, err)
		return err
	}

	if cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !sets.NewString(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
			cluster.ObjectMeta.Finalizers = append(cluster.ObjectMeta.Finalizers, clusterv1alpha1.Finalizer)
			if cluster, err = c.clusterClient.Update(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if sets.NewString(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {
			// remove our cluster finalizer
			finalizers := sets.NewString(cluster.ObjectMeta.Finalizers...)
			finalizers.Delete(clusterv1alpha1.Finalizer)
			cluster.ObjectMeta.Finalizers = finalizers.List()
			if _, err = c.clusterClient.Update(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
		return nil
	}

	// save a old copy of cluster
	oldCluster := cluster.DeepCopy()

	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		klog.V(5).Infof("Skipping to join cluster %s cause the kubeconfig is empty", cluster.Name)
		return nil
	}

	// build up cached cluster data if there isn't any
	c.mu.Lock()
	clusterDt, ok := c.clusterMap[cluster.Name]
	if !ok || clusterDt == nil || !equality.Semantic.DeepEqual(clusterDt.cachedKubeconfig, cluster.Spec.Connection.KubeConfig) {
		clusterDt, err = buildClusterData(cluster.Spec.Connection.KubeConfig)
		if err != nil {
			c.mu.Unlock()
			return err
		}
		c.clusterMap[cluster.Name] = clusterDt
	}
	c.mu.Unlock()

	// TODO: add cluster federation

	// cluster is ready, we can pull kubernetes cluster info through agent
	// since there is no agent necessary for host cluster, so updates for host cluster
	// is safe.
	if len(cluster.Spec.Connection.KubernetesAPIEndpoint) == 0 {
		cluster.Spec.Connection.KubernetesAPIEndpoint = clusterDt.config.Host
	}

	version, err := clusterDt.client.Discovery().ServerVersion()
	if err != nil {
		klog.Errorf("Failed to get kubernetes version, %#v", err)
		return err
	}

	cluster.Status.KubernetesVersion = version.GitVersion

	nodes, err := clusterDt.client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get cluster nodes, %#v", err)
		return err
	}

	cluster.Status.NodeCount = len(nodes.Items)

	configz, err := c.tryToFetchCaptainComponents(clusterDt.config.Host, clusterDt.transport)
	if err == nil {
		cluster.Status.Configz = configz
	}

	v, err := c.tryFetchCaptainVersion(clusterDt.config.Host, clusterDt.transport)
	if err != nil {
		klog.Errorf("failed to get Captain version, err: %#v", err)
	} else {
		cluster.Status.CaptainVersion = v
	}

	// label cluster host cluster if configz["multicluster"]==true
	if mc, ok := configz[configzMultiCluster]; ok && mc && c.checkIfClusterIsHostCluster(nodes) {
		if cluster.Labels == nil {
			cluster.Labels = make(map[string]string)
		}
		cluster.Labels[clusterv1alpha1.HostCluster] = ""
	}

	readyConditon := clusterv1alpha1.ClusterCondition{
		Type:               clusterv1alpha1.ClusterReady,
		Status:             v1.ConditionTrue,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             string(clusterv1alpha1.ClusterReady),
		Message:            "Cluster is available now",
	}
	c.updateClusterCondition(cluster, readyConditon)

	if !reflect.DeepEqual(oldCluster, cluster) {
		_, err = c.clusterClient.Update(context.TODO(), cluster, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Failed to update cluster status, %#v", err)
			return err
		}
	}

	return nil
}

func (c *clusterController) checkIfClusterIsHostCluster(memberClusterNodes *v1.NodeList) bool {
	hostNodes, err := c.client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false
	}

	if hostNodes == nil || memberClusterNodes == nil {
		return false
	}

	if len(hostNodes.Items) != len(memberClusterNodes.Items) {
		return false
	}

	if len(hostNodes.Items) > 0 && (hostNodes.Items[0].Status.NodeInfo.MachineID != memberClusterNodes.Items[0].Status.NodeInfo.MachineID) {
		return false
	}

	return true
}

// tryToFetchCaptainComponents will send requests to member cluster configz api using kube-apiserver proxy way
func (c *clusterController) tryToFetchCaptainComponents(host string, transport http.RoundTripper) (map[string]bool, error) {
	client := http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	response, err := client.Get(fmt.Sprintf(proxyFormat, host, "capis/config.captain.io/v1alpha2/configs/configz"))
	if err != nil {
		klog.V(4).Infof("Failed to get captain components, error %v", err)
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		klog.V(4).Infof("Response status code isn't 200.")
		return nil, fmt.Errorf("response code %d", response.StatusCode)
	}

	configz := make(map[string]bool)
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&configz)
	if err != nil {
		klog.V(4).Infof("Decode error %v", err)
		return nil, err
	}
	return configz, nil
}

//
func (c *clusterController) tryFetchCaptainVersion(host string, transport http.RoundTripper) (string, error) {
	client := http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	response, err := client.Get(fmt.Sprintf(proxyFormat, host, "capis/version"))
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		klog.V(4).Infof("Response status code isn't 200.")
		return "", fmt.Errorf("response code %d", response.StatusCode)
	}

	info := version.Info{}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&info)
	if err != nil {
		return "", err
	}

	// currently, we captain v2.1 can not be joined as a member cluster and it will never be reconciled,
	// so we don't consider that situation
	// for captain v3.0.0, the gitVersion is always v0.0.0, so we return v3.0.0
	if info.GitVersion == "v0.0.0" {
		return "v3.0.0", nil
	}

	if len(info.GitVersion) == 0 {
		return "unknown", nil
	}

	return info.GitVersion, nil
}

func (c *clusterController) addCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get cluster key %s failed", cluster.Name))
		return
	}

	c.queue.Add(key)
}

func (c *clusterController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetries {
		klog.V(2).Infof("Error syncing cluster %s, retrying, %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	klog.V(4).Infof("Dropping cluster %s out of the queue.", key)
	c.queue.Forget(key)
	utilruntime.HandleError(err)
}

// isConditionTrue checks cluster specific condition value is True, return false if condition not exists
func isConditionTrue(cluster *clusterv1alpha1.Cluster, conditionType clusterv1alpha1.ClusterConditionType) bool {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == conditionType && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// updateClusterCondition updates condition in cluster conditions using giving condition
// adds condition if not existed
func (c *clusterController) updateClusterCondition(cluster *clusterv1alpha1.Cluster, condition clusterv1alpha1.ClusterCondition) {
	if cluster.Status.Conditions == nil {
		cluster.Status.Conditions = make([]clusterv1alpha1.ClusterCondition, 0)
	}

	newConditions := make([]clusterv1alpha1.ClusterCondition, 0)
	for _, cond := range cluster.Status.Conditions {
		if cond.Type == condition.Type {
			continue
		}
		newConditions = append(newConditions, cond)
	}

	newConditions = append(newConditions, condition)
	cluster.Status.Conditions = newConditions
}
