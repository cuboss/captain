package storageclass

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	v1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type storageclassProvider struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) storageclassProvider {
	return storageclassProvider{informers: informer}
}

func (sc storageclassProvider) Get(_, name string) (runtime.Object, error) {
	return sc.informers.Storage().V1().StorageClasses().Lister().Get(name)

}

func (sc storageclassProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := sc.informers.Storage().V1().StorageClasses().Lister().List(query.GetSelector())
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
	storageClass, ok := object.(*v1.StorageClass)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaFilter(storageClass.ObjectMeta, filter)
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftNS, ok := left.(*v1.StorageClass)
	if !ok {
		return false
	}
	rightNS, ok := right.(*v1.StorageClass)
	if !ok {
		return false
	}
	return alpha1.DefaultObjectMetaCompare(leftNS.ObjectMeta, rightNS.ObjectMeta, field)
}
