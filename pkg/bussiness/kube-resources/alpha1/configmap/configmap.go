package configmap

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type configmapProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) configmapProvider {
	return configmapProvider{sharedInformers: informer}
}

func (cm configmapProvider) Get(namespace, name string) (runtime.Object, error) {
	return cm.sharedInformers.Core().V1().ConfigMaps().Lister().ConfigMaps(namespace).Get(name)
}

func (cm configmapProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := cm.sharedInformers.Core().V1().ConfigMaps().Lister().ConfigMaps(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, configMap := range raw {
		result = append(result, configMap)
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	configMap, ok := object.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaFilter(configMap.ObjectMeta, filter)
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftCM, ok := left.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	rightCM, ok := right.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaCompare(leftCM.ObjectMeta, rightCM.ObjectMeta, field)
}
