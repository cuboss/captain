package resource

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/bussiness/kube-resources/alpha1/clusterrole"
	"captain/pkg/bussiness/kube-resources/alpha1/clusterrolebinding"
	"captain/pkg/bussiness/kube-resources/alpha1/configmap"
	"captain/pkg/bussiness/kube-resources/alpha1/cronjob"
	"captain/pkg/bussiness/kube-resources/alpha1/daemonset"
	"captain/pkg/bussiness/kube-resources/alpha1/deployment"
	"captain/pkg/bussiness/kube-resources/alpha1/ingress"
	"captain/pkg/bussiness/kube-resources/alpha1/job"
	"captain/pkg/bussiness/kube-resources/alpha1/namespace"
	"captain/pkg/bussiness/kube-resources/alpha1/networkpolicy"
	"captain/pkg/bussiness/kube-resources/alpha1/node"
	"captain/pkg/bussiness/kube-resources/alpha1/persistentvolume"
	"captain/pkg/bussiness/kube-resources/alpha1/persistentvolumeclaim"
	"captain/pkg/bussiness/kube-resources/alpha1/pod"
	"captain/pkg/bussiness/kube-resources/alpha1/role"
	"captain/pkg/bussiness/kube-resources/alpha1/rolebinding"
	"captain/pkg/bussiness/kube-resources/alpha1/secret"
	"captain/pkg/bussiness/kube-resources/alpha1/service"
	"captain/pkg/bussiness/kube-resources/alpha1/serviceaccount"
	"captain/pkg/bussiness/kube-resources/alpha1/statefulset"
	"captain/pkg/bussiness/kube-resources/alpha1/storageclass"
	"captain/pkg/informers"
	"captain/pkg/server/config"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"captain/pkg/utils/clusterclient"
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

var (
	NamespaceGVR             = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	NodeGVR                  = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	ClusterroleGVR           = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}
	StorageclassGVR          = schema.GroupVersionResource{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"}
	PersistentvolumeGVR      = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumes"}
	DeploymentGVR            = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	StatefulsetGVR           = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	PodGVR                   = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	JobGVR                   = schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	CronJobGVR               = schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjobs"}
	CronJobBatchV1beta1GVR   = schema.GroupVersionResource{Group: "batch", Version: "v1beta1", Resource: "cronjobsv1beta1"}
	DaemonsetGVR             = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
	IngresseGVR              = schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
	IngresseV1beta1GVR       = schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "ingressesv1beta1"}
	ServiceGVR               = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	ConfigmapGVR             = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	PersistentvolumeClaimGVR = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaims"}
	SecretGVR                = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	ServiceaccountGVR        = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccounts"}
	RoleGVR                  = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}
	ClusterrolebindingGVR    = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}
	RolebindingGVR           = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}
	NetworkpolicieGVR        = schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"}
	ErrResourceNotSupported  = errors.New("resource is not supported")
)

// ResourceProcessor ... processing resources including kube-native, sevice mesh , others kinds of cloud-native resources
type ResourceProcessor struct {
	clusterResourceProcessors    map[schema.GroupVersionResource]alpha1.KubeResProvider
	namespacedResourceProcessors map[schema.GroupVersionResource]alpha1.KubeResProvider

	multiClusterResourceProcessors map[schema.GroupVersionResource]alpha1.MultiClusterKubeResProvider
}

func NewResourceProcessor(factory informers.CapInformerFactory, cache cache.Cache, config *config.Config) *ResourceProcessor {
	namespacedResourceProcessors := make(map[schema.GroupVersionResource]alpha1.KubeResProvider)
	clusterResourceProcessors := make(map[schema.GroupVersionResource]alpha1.KubeResProvider)

	//native kube resources
	clusterResourceProcessors[NamespaceGVR] = namespace.New(factory.KubernetesSharedInformerFactory())
	clusterResourceProcessors[NodeGVR] = node.New(factory.KubernetesSharedInformerFactory())
	clusterResourceProcessors[ClusterroleGVR] = clusterrole.New(factory.KubernetesSharedInformerFactory())
	clusterResourceProcessors[StorageclassGVR] = storageclass.New(factory.KubernetesSharedInformerFactory())
	clusterResourceProcessors[PersistentvolumeGVR] = persistentvolume.New(factory.KubernetesSharedInformerFactory())
	clusterResourceProcessors[ClusterrolebindingGVR] = clusterrolebinding.New(factory.KubernetesSharedInformerFactory())

	namespacedResourceProcessors[DeploymentGVR] = deployment.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[PodGVR] = pod.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[StatefulsetGVR] = statefulset.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[JobGVR] = job.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[CronJobGVR] = cronjob.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[CronJobBatchV1beta1GVR] = cronjob.NewBatchV1beta1(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[DaemonsetGVR] = daemonset.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[IngresseGVR] = ingress.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[IngresseV1beta1GVR] = ingress.NewV1beta1IngressProvider(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[ServiceGVR] = service.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[ConfigmapGVR] = configmap.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[PersistentvolumeClaimGVR] = persistentvolumeclaim.New(factory.KubernetesSharedInformerFactory(), factory.SnapshotSharedInformerFactory())
	namespacedResourceProcessors[SecretGVR] = secret.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[RolebindingGVR] = rolebinding.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[RoleGVR] = role.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[ServiceaccountGVR] = serviceaccount.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceProcessors[NetworkpolicieGVR] = networkpolicy.New(factory.KubernetesSharedInformerFactory())

	// multi cluster native kube resource
	multiClusterResourceProcessors := make(map[schema.GroupVersionResource]alpha1.MultiClusterKubeResProvider)
	clients := clusterclient.NewClusterClients(factory.CaptainSharedInformerFactory().Cluster().V1alpha1().Clusters(), config.MultiClusterOptions)
	multiClusterResourceProcessors[NamespaceGVR] = namespace.NewMCResProvider(clients)
	multiClusterResourceProcessors[NodeGVR] = node.NewMCResProvider(clients)
	multiClusterResourceProcessors[ClusterroleGVR] = clusterrole.NewMCResProvider(clients)
	multiClusterResourceProcessors[StorageclassGVR] = storageclass.NewMCResProvider(clients)
	multiClusterResourceProcessors[PersistentvolumeGVR] = persistentvolume.NewMCResProvider(clients)
	multiClusterResourceProcessors[ClusterrolebindingGVR] = clusterrolebinding.NewMCResProvider(clients)

	multiClusterResourceProcessors[DeploymentGVR] = deployment.NewMCResProvider(clients)
	multiClusterResourceProcessors[PodGVR] = pod.NewMCResProvider(clients)
	multiClusterResourceProcessors[StatefulsetGVR] = statefulset.NewMCResProvider(clients)
	multiClusterResourceProcessors[JobGVR] = job.NewMCResProvider(clients)
	multiClusterResourceProcessors[CronJobGVR] = cronjob.NewMCResProvider(clients)
	multiClusterResourceProcessors[CronJobBatchV1beta1GVR] = cronjob.NewMCBatchV1beta1ResProvider(clients)
	multiClusterResourceProcessors[DaemonsetGVR] = daemonset.NewMCResProvider(clients)
	multiClusterResourceProcessors[IngresseGVR] = ingress.NewMCResProvider(clients)
	multiClusterResourceProcessors[IngresseV1beta1GVR] = ingress.NewMCV1beta1ResProvider(clients)
	multiClusterResourceProcessors[ServiceGVR] = service.NewMCResProvider(clients)
	multiClusterResourceProcessors[ConfigmapGVR] = configmap.NewMCResProvider(clients)
	multiClusterResourceProcessors[PersistentvolumeClaimGVR] = persistentvolumeclaim.NewMCResProvider(clients)
	multiClusterResourceProcessors[SecretGVR] = secret.NewMCResProvider(clients)
	multiClusterResourceProcessors[RolebindingGVR] = rolebinding.NewMCResProvider(clients)
	multiClusterResourceProcessors[RoleGVR] = role.NewMCResProvider(clients)
	multiClusterResourceProcessors[ServiceaccountGVR] = serviceaccount.NewMCResProvider(clients)
	multiClusterResourceProcessors[NetworkpolicieGVR] = networkpolicy.NewMCResProvider(clients)

	return &ResourceProcessor{
		namespacedResourceProcessors:   namespacedResourceProcessors,
		clusterResourceProcessors:      clusterResourceProcessors,
		multiClusterResourceProcessors: multiClusterResourceProcessors,
	}
}

// TryResource will retrieve a getter with resource name, it doesn't guarantee find resource with correct group version
// need to refactor this use schema.GroupVersionResource
func (r *ResourceProcessor) TryResource(clusterScope bool, resource string) alpha1.KubeResProvider {
	if clusterScope {
		for k, v := range r.clusterResourceProcessors {
			if k.Resource == resource {
				return v
			}
		}
	}
	for k, v := range r.namespacedResourceProcessors {
		if k.Resource == resource {
			return v
		}
	}

	return nil
}

func (r *ResourceProcessor) TryMultiClusterResource(resource string) alpha1.MultiClusterKubeResProvider {
	for k, v := range r.multiClusterResourceProcessors {
		if k.Resource == resource {
			return v
		}
	}
	return nil
}

func (r *ResourceProcessor) Get(region, cluster, resource, namespace, name string) (runtime.Object, error) {
	if alpha1.IsHostCluster(region, cluster) {
		clusterScope := namespace == ""
		getter := r.TryResource(clusterScope, resource)
		if getter == nil {
			return nil, ErrResourceNotSupported
		}
		return getter.Get(namespace, name)
	}
	getter := r.TryMultiClusterResource(resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.Get(region, cluster, namespace, name)
}

func (r *ResourceProcessor) List(region, cluster, resource, namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	if alpha1.IsHostCluster(region, cluster) {
		// parse cluster scope or not
		clusterScope := namespace == ""

		provider := r.TryResource(clusterScope, resource)
		if provider == nil {
			return nil, ErrResourceNotSupported
		}
		return provider.List(namespace, query)
	}
	provider := r.TryMultiClusterResource(resource)
	if provider == nil {
		return nil, ErrResourceNotSupported
	}
	return provider.List(region, cluster, namespace, query)
}
