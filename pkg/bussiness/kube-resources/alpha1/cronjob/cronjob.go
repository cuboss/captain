package cronjob

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"strings"

	"k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

const (
	LastScheduleTime = "lastScheduleTime"
	StatusPaused     = "paused"
	StatusRunning    = "running"
)

type cronjobProvider struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) cronjobProvider {
	return cronjobProvider{informers: informer}
}

func (cj cronjobProvider) Get(namespace, name string) (runtime.Object, error) {
	return cj.informers.Batch().V1().CronJobs().Lister().CronJobs(namespace).Get(name)

}

func (cj cronjobProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := cj.informers.Batch().V1().CronJobs().Lister().CronJobs(namespace).List(query.GetSelector())
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
	cronJob, ok := object.(*v1beta1.CronJob)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(cronJobStatus(cronJob), string(filter.Value)) == 0
	default:
		return alpha1.DefaultObjectMetaFilter(cronJob.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftcj, ok := left.(*v1beta1.CronJob)
	if !ok {
		return false
	}
	rightcj, ok := right.(*v1beta1.CronJob)
	if !ok {
		return false
	}
	switch field {
	case LastScheduleTime:
		if leftcj.Status.LastScheduleTime == nil {
			return true
		}
		if rightcj.Status.LastScheduleTime == nil {
			return false
		}
		return leftcj.Status.LastScheduleTime.Before(rightcj.Status.LastScheduleTime)
	default:
		return alpha1.DefaultObjectMetaCompare(leftcj.ObjectMeta, rightcj.ObjectMeta, field)
	}
}

func cronJobStatus(item *v1beta1.CronJob) string {
	if item.Spec.Suspend != nil && *item.Spec.Suspend {
		return StatusPaused
	}
	return StatusRunning
}
