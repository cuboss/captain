package persistentvolume

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"strings"
)

const (
	storageClassName = "storageClassName"
)

type persistentvolumeProvider struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) persistentvolumeProvider {
	return persistentvolumeProvider{informers: informer}
}

func (pv persistentvolumeProvider) Get(_, name string) (runtime.Object, error) {
	return pv.informers.Core().V1().PersistentVolumes().Lister().Get(name)

}

func (pv persistentvolumeProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := pv.informers.Core().V1().PersistentVolumes().Lister().List(query.GetSelector())
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
	persistentVolume, ok := object.(*corev1.PersistentVolume)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.EqualFold(string(persistentVolume.Status.Phase), string(filter.Value))
	case storageClassName:
		return persistentVolume.Spec.StorageClassName != "" && persistentVolume.Spec.StorageClassName == string(filter.Value)

	default:
		return alpha1.DefaultObjectMetaFilter(persistentVolume.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	pv1, ok := left.(*corev1.PersistentVolume)
	if !ok {
		return false
	}
	pv2, ok := right.(*corev1.PersistentVolume)
	if !ok {
		return false
	}
	return alpha1.DefaultObjectMetaCompare(pv1.ObjectMeta, pv2.ObjectMeta, field)

}
