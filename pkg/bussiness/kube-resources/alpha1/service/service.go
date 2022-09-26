package service

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type serviceProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) serviceProvider {
	return serviceProvider{sharedInformers: informer}
}

func (svc serviceProvider) Get(namespace, name string) (runtime.Object, error) {
	return svc.sharedInformers.Core().V1().Services().Lister().Services(namespace).Get(name)
}

func (svc serviceProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := svc.sharedInformers.Core().V1().Services().Lister().Services(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, deploy := range raw {
		result = append(result, deploy)
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	service, ok := object.(*corev1.Service)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaFilter(service.ObjectMeta, filter)
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftService, ok := left.(*corev1.Service)
	if !ok {
		return false
	}

	rightService, ok := right.(*corev1.Service)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaCompare(leftService.ObjectMeta, rightService.ObjectMeta, field)
}
