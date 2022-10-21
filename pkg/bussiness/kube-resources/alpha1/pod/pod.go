package pod

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

const (
	fieldNodeName    = "nodeName"
	fieldPVCName     = "pvcName"
	fieldServiceName = "serviceName"
	fieldStatus      = "status"
)

type podProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) podProvider {
	return podProvider{sharedInformers: informer}
}

func (pd podProvider) Get(namespace, name string) (runtime.Object, error) {
	return pd.sharedInformers.Core().V1().Pods().Lister().Pods(namespace).Get(name)
}

func (pd podProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := pd.sharedInformers.Core().V1().Pods().Lister().Pods(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, pod := range raw {
		result = append(result, pod)
	}

	return alpha1.DefaultList(result, query, compareFunc, pd.filter), nil
}

func (pd *podProvider) filter(object runtime.Object, filter query.Filter) bool {
	pod, ok := object.(*v1.Pod)

	if !ok {
		return false
	}
	switch filter.Field {
	case fieldNodeName:
		return pod.Spec.NodeName == string(filter.Value)
	case fieldPVCName:
		return podBindPVC(pod, string(filter.Value))
	case fieldServiceName:
		return pd.podBelongToService(pod, string(filter.Value))
	case fieldStatus:
		return string(pod.Status.Phase) == string(filter.Value)
	default:
		return alpha1.DefaultObjectMetaFilter(pod.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftPod, ok := left.(*v1.Pod)
	if !ok {
		return false
	}
	rightPod, ok := right.(*v1.Pod)
	if !ok {
		return false
	}
	return alpha1.DefaultObjectMetaCompare(leftPod.ObjectMeta, rightPod.ObjectMeta, field)

}

func (pd *podProvider) podBelongToService(item *v1.Pod, serviceName string) bool {
	service, err := pd.sharedInformers.Core().V1().Services().Lister().Services(item.Namespace).Get(serviceName)
	if err != nil {
		return false
	}

	selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
	if selector.Empty() || !selector.Matches(labels.Set(item.Labels)) {
		return false
	}
	return true
}

func podBindPVC(item *v1.Pod, pvcName string) bool {
	for _, v := range item.Spec.Volumes {
		if v.VolumeSource.PersistentVolumeClaim != nil &&
			v.VolumeSource.PersistentVolumeClaim.ClaimName == pvcName {
			return true
		}
	}
	return false
}


