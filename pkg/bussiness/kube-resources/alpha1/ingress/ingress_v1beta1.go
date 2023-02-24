package ingress

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	v1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type ingressV1beta1Provider struct {
	sharedInformers informers.SharedInformerFactory
}

func NewV1beta1IngressProvider(informer informers.SharedInformerFactory) ingressV1beta1Provider {
	return ingressV1beta1Provider{sharedInformers: informer}
}

func (ing ingressV1beta1Provider) Get(namespace, name string) (runtime.Object, error) {
	return ing.sharedInformers.Networking().V1beta1().Ingresses().Lister().Ingresses(namespace).Get(name)
}

func (ing ingressV1beta1Provider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := ing.sharedInformers.Networking().V1beta1().Ingresses().Lister().Ingresses(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, deploy := range raw {
		result = append(result, deploy)
	}

	return alpha1.DefaultList(result, query, v1beta1CompareFunc, v1beta1Filter), nil
}

func v1beta1Filter(object runtime.Object, filter query.Filter) bool {
	ingress, ok := object.(*v1beta1.Ingress)
	if !ok {
		return false
	}
	return alpha1.DefaultObjectMetaFilter(ingress.ObjectMeta, filter)
}

func v1beta1CompareFunc(left, right runtime.Object, field query.Field) bool {

	leftIngress, ok := left.(*v1beta1.Ingress)
	if !ok {
		return false
	}

	rightIngress, ok := right.(*v1beta1.Ingress)
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
