package job

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"strings"
	"time"
)

const (
	jobFailed    = "failed"
	jobCompleted = "completed"
	jobRunning   = "running"
)

type jobProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) jobProvider {
	return jobProvider{sharedInformers: informer}
}

func (j jobProvider) Get(namespace, name string) (runtime.Object, error) {
	return j.sharedInformers.Batch().V1().Jobs().Lister().Jobs(namespace).Get(name)
}

func (j jobProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := j.sharedInformers.Batch().V1().Jobs().Lister().Jobs(namespace).List(query.GetSelector())
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
	job, ok := object.(*batchv1.Job)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(jobStatus(job.Status), string(filter.Value)) == 0
	default:
		return alpha1.DefaultObjectMetaFilter(job.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftJob, ok := left.(*batchv1.Job)
	if !ok {
		return false
	}

	rightJob, ok := right.(*batchv1.Job)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftJob).After(lastUpdateTime(rightJob))
	case query.FieldStatus:
		return strings.Compare(jobStatus(leftJob.Status), jobStatus(rightJob.Status)) > 0
	default:
		return alpha1.DefaultObjectMetaCompare(leftJob.ObjectMeta, rightJob.ObjectMeta, field)
	}

}

func jobStatus(status batchv1.JobStatus) string {
	for _, condition := range status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			return jobCompleted
		} else if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			return jobFailed
		}
	}

	return jobRunning
}

func lastUpdateTime(job *batchv1.Job) time.Time {
	lut := job.CreationTimestamp.Time
	for _, condition := range job.Status.Conditions {
		if condition.LastTransitionTime.After(lut) {
			lut = condition.LastTransitionTime.Time
		}
	}
	return lut
}
