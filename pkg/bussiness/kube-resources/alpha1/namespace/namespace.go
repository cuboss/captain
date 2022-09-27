package namespace

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type namespaceProvider struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) namespaceProvider {
	return namespaceProvider{informers: informer}
}

func (ns namespaceProvider) Get(_, name string) (runtime.Object, error) {
	return ns.informers.Core().V1().Namespaces().Lister().Get(name)

}

func (ns namespaceProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := ns.informers.Core().V1().Namespaces().Lister().List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, nasp := range raw {
		result = append(result, nasp)
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	namespace, ok := object.(*v1.Namespace)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(string(namespace.Status.Phase), string(filter.Value)) == 0
	default:
		return alpha1.DefaultObjectMetaFilter(namespace.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftNS, ok := left.(*v1.Namespace)
	if !ok {
		return false
	}
	rightNS, ok := right.(*v1.Namespace)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftNS).After(lastUpdateTime(rightNS))
	default:
		return alpha1.DefaultObjectMetaCompare(leftNS.ObjectMeta, rightNS.ObjectMeta, field)
	}
}

func lastUpdateTime(namespase *v1.Namespace) time.Time {
	recent := namespase.CreationTimestamp.Time

	for _, condition := range namespase.Status.Conditions {
		if condition.LastTransitionTime.After(recent) {
			recent = condition.LastTransitionTime.Time
		}
	}
	return recent
}
