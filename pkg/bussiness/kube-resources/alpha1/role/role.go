package role

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type roleProvider struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) roleProvider {
	return roleProvider{informers: informer}
}

func (cr roleProvider) Get(namespace, name string) (runtime.Object, error) {
	return cr.informers.Rbac().V1().Roles().Lister().Roles(namespace).Get(name)

}

func (cr roleProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	var roles []*rbacv1.Role
	var err error

	roles, err = cr.informers.Rbac().V1().Roles().Lister().Roles(namespace).List(query.GetSelector())

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, role := range roles {
		result = append(result, role)
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*rbacv1.Role)

	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftRole, ok := left.(*rbacv1.Role)
	if !ok {
		return false
	}

	rightRole, ok := right.(*rbacv1.Role)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaCompare(leftRole.ObjectMeta, rightRole.ObjectMeta, field)
}
