package networkpolicy

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type networkpolicyProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) networkpolicyProvider {
	return networkpolicyProvider{sharedInformers: informer}
}

func (netp networkpolicyProvider) Get(namespace, name string) (runtime.Object, error) {
	return netp.sharedInformers.Networking().V1().NetworkPolicies().Lister().NetworkPolicies(namespace).Get(name)
}

func (netp networkpolicyProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	nps, err := netp.sharedInformers.Networking().V1().NetworkPolicies().Lister().NetworkPolicies(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, item := range nps {
		result = append(result, item)
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	np, ok := object.(*v1.NetworkPolicy)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaFilter(np.ObjectMeta, filter)
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftNP, ok := left.(*v1.NetworkPolicy)
	if !ok {
		return false
	}

	rightNP, ok := right.(*v1.NetworkPolicy)
	if !ok {
		return true
	}
	return alpha1.DefaultObjectMetaCompare(leftNP.ObjectMeta, rightNP.ObjectMeta, field)
}
