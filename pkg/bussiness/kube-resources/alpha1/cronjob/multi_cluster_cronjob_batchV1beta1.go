package cronjob

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"captain/pkg/utils/clusterclient"
)

type mcCronJobBatchV1beta1Provider struct {
	clusterclient.ClusterClients
}

func NewMCBatchV1beta1ResProvider(clients clusterclient.ClusterClients) mcCronJobBatchV1beta1Provider {
	return mcCronJobBatchV1beta1Provider{ClusterClients: clients}
}

func (pd mcCronJobBatchV1beta1Provider) Get(region, cluster, namespace, name string) (runtime.Object, error) {
	cli, err := pd.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}

	return cli.BatchV1beta1().CronJobs(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (pd mcCronJobBatchV1beta1Provider) List(region, cluster, namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	cli, err := pd.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}
	list, err := cli.BatchV1beta1().CronJobs(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: query.LabelSelector})
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	if list != nil && list.Items != nil {
		for i := 0; i < len(list.Items); i++ {
			result = append(result, &list.Items[i])
		}
	}

	return alpha1.DefaultList(result, query, v1Beta1CompareFunc, v1Beta1Filter), nil
}
