package statefulset

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"strings"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

const (
	StatusStopped  = "stopped"
	StatusRunning  = "running"
	StatusUpdating = "updating"
)

type statefulSetProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) statefulSetProvider {
	return statefulSetProvider{sharedInformers: informer}
}

func (sts statefulSetProvider) Get(namespace, name string) (runtime.Object, error) {
	return sts.sharedInformers.Apps().V1().StatefulSets().Lister().StatefulSets(namespace).Get(name)
}

func (sts statefulSetProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := sts.sharedInformers.Apps().V1().StatefulSets().Lister().StatefulSets(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, deploy := range raw {
		result = append(result, deploy)
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	statefulset, ok := object.(*v1.StatefulSet)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(statefulSetStatus(statefulset), string(filter.Value)) == 0

	default:
		return alpha1.DefaultObjectMetaFilter(statefulset.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftDeploy, ok := left.(*v1.StatefulSet)
	if !ok {
		return false
	}
	rightDeploy, ok := right.(*v1.StatefulSet)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftDeploy).After(lastUpdateTime(rightDeploy))
	default:
		return alpha1.DefaultObjectMetaCompare(leftDeploy.ObjectMeta, rightDeploy.ObjectMeta, field)
	}
}

func lastUpdateTime(statefulSet *v1.StatefulSet) time.Time {
	recent := statefulSet.CreationTimestamp.Time

	for _, condition := range statefulSet.Status.Conditions {
		if condition.LastTransitionTime.After(recent) {
			recent = condition.LastTransitionTime.Time
		}
	}
	return recent
}

func statefulSetStatus(item *v1.StatefulSet) string {
	if item.Spec.Replicas != nil {
		if item.Status.ReadyReplicas == 0 && *item.Spec.Replicas == 0 {
			return StatusStopped
		} else if item.Status.ReadyReplicas == *item.Spec.Replicas {
			return StatusRunning
		} else {
			return StatusUpdating
		}
	}
	return StatusStopped
}
