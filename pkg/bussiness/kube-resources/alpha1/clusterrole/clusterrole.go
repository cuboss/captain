package clusterrole

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"captain/pkg/utils/k8sutil"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type clusterRoleProvider struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) clusterRoleProvider {
	return clusterRoleProvider{informers: informer}
}

func (cr clusterRoleProvider) Get(_, name string) (runtime.Object, error) {
	return cr.informers.Rbac().V1().ClusterRoles().Lister().Get(name)

}

func (cr clusterRoleProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := cr.informers.Rbac().V1().ClusterRoles().Lister().List(query.GetSelector())
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
	clusterRole, ok := object.(*rbac.ClusterRole)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldOwnerKind:
		fallthrough
	case query.FieldOwnerName:
		kind := filter.Field
		name := filter.Value
		if !k8sutil.IsControlledBy(clusterRole.OwnerReferences, string(kind), string(name)) {
			return false
		}
	default:
		return alpha1.DefaultObjectMetaFilter(clusterRole.ObjectMeta, filter)
	}
	return true
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftNS, ok := left.(*rbac.ClusterRole)
	if !ok {
		return false
	}
	rightNS, ok := right.(*rbac.ClusterRole)
	if !ok {
		return false
	}
	switch field {
	default:
		return alpha1.DefaultObjectMetaCompare(leftNS.ObjectMeta, rightNS.ObjectMeta, field)
	}
}
