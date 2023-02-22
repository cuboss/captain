package resourcequota

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type resourcequotaProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) resourcequotaProvider {
	return resourcequotaProvider{sharedInformers: informer}
}

func (rq resourcequotaProvider) Get(namespace, name string) (runtime.Object, error) {
	return rq.sharedInformers.Core().V1().ResourceQuotas().Lister().ResourceQuotas(namespace).Get(name)
}

func (rq resourcequotaProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := rq.sharedInformers.Core().V1().ResourceQuotas().Lister().ResourceQuotas(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, resourceQuota := range raw {
		result = append(result, resourceQuota)
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	resourceQuota, ok := object.(*corev1.ResourceQuota)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaFilter(resourceQuota.ObjectMeta, filter)
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftrq, ok := left.(*corev1.ResourceQuota)
	if !ok {
		return false
	}

	rightrq, ok := right.(*corev1.ResourceQuota)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaCompare(leftrq.ObjectMeta, rightrq.ObjectMeta, field)
}
