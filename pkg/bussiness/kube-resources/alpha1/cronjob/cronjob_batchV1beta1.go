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

type cronjobV1beta1Provider struct {
	informers informers.SharedInformerFactory
}

func NewBatchV1beta1(informer informers.SharedInformerFactory) cronjobV1beta1Provider {
	return cronjobV1beta1Provider{informers: informer}
}

func (cj cronjobV1beta1Provider) Get(namespace, name string) (runtime.Object, error) {
	return cj.informers.Batch().V1beta1().CronJobs().Lister().CronJobs(namespace).Get(name)

}

func (cj cronjobV1beta1Provider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := cj.informers.Batch().V1beta1().CronJobs().Lister().CronJobs(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, nasp := range raw {
		result = append(result, nasp)
	}

	return alpha1.DefaultList(result, query, v1Beta1CompareFunc, v1Beta1Filter), nil
}

func v1Beta1Filter(object runtime.Object, filter query.Filter) bool {
	cronJob, ok := object.(*v1beta1.CronJob)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(v1Beta1CronJobStatus(cronJob), string(filter.Value)) == 0
	default:
		return alpha1.DefaultObjectMetaFilter(cronJob.ObjectMeta, filter)
	}
}

func v1Beta1CompareFunc(left, right runtime.Object, field query.Field) bool {

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

func v1Beta1CronJobStatus(item *v1beta1.CronJob) string {
	if item.Spec.Suspend != nil && *item.Spec.Suspend {
		return StatusPaused
	}
	return StatusRunning
}
