package pod

import (
	"context"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"captain/pkg/utils/clusterclient"
)

type mcPodProvider struct {
	clusterclient.ClusterClients
}

func NewMCResProvider(clients clusterclient.ClusterClients) mcPodProvider {
	return mcPodProvider{ClusterClients: clients}
}

func (pd mcPodProvider) Get(region, cluster, namespace, name string) (runtime.Object, error) {
	cli, err := pd.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}

	return cli.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (pd mcPodProvider) List(region, cluster, namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	cli, err := pd.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}
	list, err := cli.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: query.LabelSelector})
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	if list != nil && list.Items != nil {
		for i := 0; i < len(list.Items); i++ {
			result = append(result, &list.Items[i])
		}
	}

	podCli := PodProviderClient{Clientset: cli}
	return alpha1.DefaultList(result, query, compareFunc, podCli.filter), nil
}

type PodProviderClient struct {
	*kubernetes.Clientset
	replicaSets *appv1.ReplicaSetList
	service     *v1.Service
}

func (c *PodProviderClient) filter(object runtime.Object, filter query.Filter) bool {
	pod, ok := object.(*v1.Pod)
	if !ok {
		return false
	}
	if !ok {
		return false
	}
	switch filter.Field {
	case fieldNodeName:
		return pod.Spec.NodeName == string(filter.Value)
	case fieldPVCName:
		return podBindPVC(pod, string(filter.Value))
	case fieldServiceName:
		return c.podBelongToService(pod, string(filter.Value))
	case fieldStatus:
		return string(pod.Status.Phase) == string(filter.Value)
	default:
		return alpha1.DefaultObjectMetaFilter(pod.ObjectMeta, filter)
	}
}

func (c *PodProviderClient) podBelongToService(item *v1.Pod, serviceName string) bool {
	var service *v1.Service
	if c.service != nil {
		service = c.service
	} else {
		service, err := c.CoreV1().Services(item.Namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
		if err != nil {
			return false
		}
		c.service = service
	}
	if service != nil {
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Empty() || !selector.Matches(labels.Set(item.Labels)) {
			return false
		}
	}
	return true
}
