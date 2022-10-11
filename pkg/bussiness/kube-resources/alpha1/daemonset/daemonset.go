package daemonset

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"strings"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type daemonsetProvider struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) daemonsetProvider {
	return daemonsetProvider{informers: informer}
}

func (dms daemonsetProvider) Get(namespace, name string) (runtime.Object, error) {
	return dms.informers.Apps().V1().DaemonSets().Lister().DaemonSets(namespace).Get(name)

}

func (dms daemonsetProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := dms.informers.Apps().V1().DaemonSets().Lister().DaemonSets(namespace).List(query.GetSelector())
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
	daemonSet, ok := object.(*appsv1.DaemonSet)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(daemonsetStatus(&daemonSet.Status), string(filter.Value)) == 0
	default:
		return alpha1.DefaultObjectMetaFilter(daemonSet.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftDaemonSet, ok := left.(*appsv1.DaemonSet)
	if !ok {
		return false
	}

	rightDaemonSet, ok := right.(*appsv1.DaemonSet)
	if !ok {
		return false
	}

	return alpha1.DefaultObjectMetaCompare(leftDaemonSet.ObjectMeta, rightDaemonSet.ObjectMeta, field)
}

func daemonsetStatus(status *appsv1.DaemonSetStatus) string {
	if status.DesiredNumberScheduled == 0 && status.NumberReady == 0 {
		return statusStopped
	} else if status.DesiredNumberScheduled == status.NumberReady {
		return statusRunning
	} else {
		return statusUpdating
	}
}
