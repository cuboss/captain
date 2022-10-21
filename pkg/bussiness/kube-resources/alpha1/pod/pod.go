package pod

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"strings"
)

const (
	fieldNodeName    = "nodeName"
	fieldPVCName     = "pvcName"
	fieldServiceName = "serviceName"
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
	case query.FieldOwner:
		kn := strings.Split(string(filter.Value), "=")
		if len(kn) != 2 {
			return false
		}
		kind := kn[0]
		name := kn[1]
		return pd.podBelongTo(pod, kind, name)
	case fieldNodeName:
		return pod.Spec.NodeName == string(filter.Value)
	case fieldPVCName:
		return podBindPVC(pod, string(filter.Value))
	case fieldServiceName:
		return pd.podBelongToService(pod, string(filter.Value))
	case query.FieldStatus:
		return strings.Compare(string(pod.Status.Phase), string(filter.Value)) == 0
	default:
		return alpha1.DefaultObjectMetaFilter(pod.ObjectMeta, filter)
	}
	return false
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
	switch field {
	case query.FieldStartTime:
		if leftPod.Status.StartTime == nil {
			return false
		}
		if rightPod.Status.StartTime == nil {
			return true
		}
	default:
		return alpha1.DefaultObjectMetaCompare(leftPod.ObjectMeta, rightPod.ObjectMeta, field)
	}
	return false
}
func (pd *podProvider) podBelongTo(item *v1.Pod, kind string, name string) bool {
	switch kind {
	case "Deployment":
		if pd.podBelongToDeployment(item, name) {
			return true
		}
	case "ReplicaSet":
		if podBelongToReplicaSet(item, name) {
			return true
		}
	case "DaemonSet":
		if podBelongToDaemonSet(item, name) {
			return true
		}
	case "StatefulSet":
		if podBelongToStatefulSet(item, name) {
			return true
		}
	case "Job":
		if podBelongToJob(item, name) {
			return true
		}
	}
	return false
}

func replicaSetBelongToDeployment(replicaSet *appsv1.ReplicaSet, deploymentName string) bool {
	for _, owner := range replicaSet.OwnerReferences {
		if owner.Kind == "Deployment" && owner.Name == deploymentName {
			return true
		}
	}
	return false
}

func podBelongToDaemonSet(item *v1.Pod, name string) bool {
	for _, owner := range item.OwnerReferences {
		if owner.Kind == "DaemonSet" && owner.Name == name {
			return true
		}
	}
	return false
}

func podBelongToJob(item *v1.Pod, name string) bool {
	for _, owner := range item.OwnerReferences {
		if owner.Kind == "Job" && owner.Name == name {
			return true
		}
	}
	return false
}

func podBelongToReplicaSet(item *v1.Pod, replicaSetName string) bool {
	for _, owner := range item.OwnerReferences {
		if owner.Kind == "ReplicaSet" && owner.Name == replicaSetName {
			return true
		}
	}
	return false
}

func podBelongToStatefulSet(item *v1.Pod, statefulSetName string) bool {
	for _, owner := range item.OwnerReferences {
		if owner.Kind == "StatefulSet" && owner.Name == statefulSetName {
			return true
		}
	}
	return false
}

func (pd *podProvider) podBelongToDeployment(item *v1.Pod, deploymentName string) bool {
	replicas, err := pd.sharedInformers.Apps().V1().ReplicaSets().Lister().ReplicaSets(item.Namespace).List(labels.Everything())
	if err != nil {
		return false
	}

	for _, r := range replicas {
		if replicaSetBelongToDeployment(r, deploymentName) && podBelongToReplicaSet(item, r.Name) {
			return true
		}
	}

	return false
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
