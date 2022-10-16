package serviceaccount

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"captain/pkg/utils/clusterclient"
)

type mcServiceAccountProvider struct {
	clusterclient.ClusterClients
}

func NewMCResProvider(clients clusterclient.ClusterClients) mcServiceAccountProvider {
	return mcServiceAccountProvider{ClusterClients: clients}
}

func (pd mcServiceAccountProvider) Get(region, cluster, namespace, name string) (runtime.Object, error) {
	cli, err := pd.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}

	return cli.CoreV1().ServiceAccounts(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (pd mcServiceAccountProvider) List(region, cluster, namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	cli, err := pd.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}
	list, err := cli.CoreV1().ServiceAccounts(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: query.LabelSelector})
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	if list != nil && list.Items != nil {
		for i := 0; i < len(list.Items); i++ {
			result = append(result, &list.Items[i])
		}
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}
