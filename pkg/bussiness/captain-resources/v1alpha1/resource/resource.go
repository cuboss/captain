package resource

import (
	"errors"

	"captain/pkg/bussiness/captain-resources/v1alpha1"
	"captain/pkg/bussiness/captain-resources/v1alpha1/cluster"
	"captain/pkg/crd"
	"captain/pkg/informers"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

var (
	ClusterGVR              = schema.GroupVersionResource{Group: "captain.io", Version: "v1alpah1", Resource: "clusters"}
	ErrResourceNotSupported = errors.New("resource is not supported")
)

// ResourceProcessor ... processing resources including kube-native, sevice mesh , others kinds of cloud-native resources
type ResourceProcessor struct {
	clusterResourceProcessors    map[schema.GroupVersionResource]v1alpha1.CaptainResProvider
	namespacedResourceProcessors map[schema.GroupVersionResource]v1alpha1.CaptainResProvider
}

func NewResourceProcessor(factory informers.CapInformerFactory, crd crd.CrdInterface, cache cache.Cache) *ResourceProcessor {
	namespacedResourceProcessors := make(map[schema.GroupVersionResource]v1alpha1.CaptainResProvider)
	clusterResourceProcessors := make(map[schema.GroupVersionResource]v1alpha1.CaptainResProvider)

	//native kube resources
	namespacedResourceProcessors[ClusterGVR] = cluster.New(factory.CaptainSharedInformerFactory(), crd)

	return &ResourceProcessor{
		namespacedResourceProcessors: namespacedResourceProcessors,
		clusterResourceProcessors:    clusterResourceProcessors,
	}
}

// TryResource will retrieve a getter with resource name, it doesn't guarantee find resource with correct group version
// need to refactor this use schema.GroupVersionResource
func (r *ResourceProcessor) TryResource(clusterScope bool, resource string) v1alpha1.CaptainResProvider {
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

func (r *ResourceProcessor) Get(resource, namespace, name string) (runtime.Object, error) {
	clusterScope := namespace == ""
	getter := r.TryResource(clusterScope, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.Get(namespace, name)
}

func (r *ResourceProcessor) List(resource, namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	// parse cluster scope or not
	clusterScope := namespace == ""

	provider := r.TryResource(clusterScope, resource)
	if provider == nil {
		return nil, ErrResourceNotSupported
	}
	return provider.List(namespace, query)
}

func (r *ResourceProcessor) Create(resource, namespace string, obj runtime.Object) (runtime.Object, error) {
	clusterScope := namespace == ""
	provider := r.TryResource(clusterScope, resource)
	if provider == nil {
		return nil, ErrResourceNotSupported
	}
	return provider.Create(namespace, obj)
}

func (r *ResourceProcessor) Delete(resource, namespace, name string) error {
	clusterScope := namespace == ""
	provider := r.TryResource(clusterScope, resource)
	if provider == nil {
		return ErrResourceNotSupported
	}
	return provider.Delete(namespace, name)
}

func (r *ResourceProcessor) Update(resource, namespace string, obj runtime.Object) (runtime.Object, error) {
	clusterScope := namespace == ""
	provider := r.TryResource(clusterScope, resource)
	if provider == nil {
		return nil, ErrResourceNotSupported
	}
	return provider.Update(namespace, obj)
}
