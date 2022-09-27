package serviceaccount

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type serviceaccountProvider struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) serviceaccountProvider {
	return serviceaccountProvider{informers: informer}
}

func (cr serviceaccountProvider) Get(namespace, name string) (runtime.Object, error) {
	return cr.informers.Core().V1().ServiceAccounts().Lister().ServiceAccounts(namespace).Get(name)

}

func (cr serviceaccountProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {

	serviceaccounts, err := cr.informers.Core().V1().ServiceAccounts().Lister().ServiceAccounts(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, serviceaccount := range serviceaccounts {
		result = append(result, serviceaccount)
	}
	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	serviceAccount, ok := object.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaFilter(serviceAccount.ObjectMeta, filter)
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftCM, ok := left.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	rightCM, ok := right.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaCompare(leftCM.ObjectMeta, rightCM.ObjectMeta, field)
}
