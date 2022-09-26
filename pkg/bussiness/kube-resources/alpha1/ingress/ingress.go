package ingress

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type ingressProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) ingressProvider {
	return ingressProvider{sharedInformers: informer}
}

func (ing ingressProvider) Get(namespace, name string) (runtime.Object, error) {
	return ing.sharedInformers.Networking().V1().Ingresses().Lister().Ingresses(namespace).Get(name)
}

func (ing ingressProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := ing.sharedInformers.Networking().V1().Ingresses().Lister().Ingresses(namespace).List(query.GetSelector())
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
	ingress, ok := object.(*v1.Ingress)
	if !ok {
		return false
	}
	return alpha1.DefaultObjectMetaFilter(ingress.ObjectMeta, filter)
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftIngress, ok := left.(*v1.Ingress)
	if !ok {
		return false
	}

	rightIngress, ok := right.(*v1.Ingress)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	default:
		return alpha1.DefaultObjectMetaCompare(leftIngress.ObjectMeta, rightIngress.ObjectMeta, field)
	}
}
