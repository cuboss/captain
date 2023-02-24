package secret

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type secretProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) secretProvider {
	return secretProvider{sharedInformers: informer}
}

func (s secretProvider) Get(namespace, name string) (runtime.Object, error) {
	return s.sharedInformers.Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
}

func (s secretProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := s.sharedInformers.Core().V1().Secrets().Lister().Secrets(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, sc := range raw {
		result = append(result, sc)
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftSecret, ok := left.(*v1.Secret)
	if !ok {
		return false
	}

	rightSecret, ok := right.(*v1.Secret)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaCompare(leftSecret.ObjectMeta, rightSecret.ObjectMeta, field)
}

func filter(object runtime.Object, filter query.Filter) bool {
	secret, ok := object.(*v1.Secret)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaFilter(secret.ObjectMeta, filter)
}
