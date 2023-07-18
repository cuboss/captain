package horizontalpodautoscaler

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"captain/pkg/utils/clusterclient"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type mcHpaV2beta2Provider struct {
	clusterclient.ClusterClients
}

func NewMCV2beta2HpaProvider(clients clusterclient.ClusterClients) mcHpaV2beta2Provider {
	return mcHpaV2beta2Provider{ClusterClients: clients}
}

func (hpa mcHpaV2beta2Provider) Get(region, cluster, namespace, name string) (runtime.Object, error) {
	cli, err := hpa.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}

	return cli.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (hpa mcHpaV2beta2Provider) List(region, cluster, namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	cli, err := hpa.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}
	list, err := cli.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: query.LabelSelector})
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	if list != nil && list.Items != nil {
		for i := 0; i < len(list.Items); i++ {
			result = append(result, &list.Items[i])
		}
	}

	return alpha1.DefaultList(result, query, v2beta2CompareFunc, v2beta2Filter), nil
}
