package limitrange

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type limitRangeProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) limitRangeProvider {
	return limitRangeProvider{sharedInformers: informer}
}

func (lr limitRangeProvider) Get(namespace, name string) (runtime.Object, error) {
	return lr.sharedInformers.Core().V1().LimitRanges().Lister().LimitRanges(namespace).Get(name)
}

func (lr limitRangeProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := lr.sharedInformers.Core().V1().LimitRanges().Lister().LimitRanges(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, limitRange := range raw {
		result = append(result, limitRange)
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	limitRange, ok := object.(*corev1.LimitRange)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaFilter(limitRange.ObjectMeta, filter)
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftlr, ok := left.(*corev1.LimitRange)
	if !ok {
		return false
	}

	rightlr, ok := right.(*corev1.LimitRange)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaCompare(leftlr.ObjectMeta, rightlr.ObjectMeta, field)
}
