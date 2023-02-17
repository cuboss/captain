package cronjob

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

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

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}
