package cluster

import (
	"context"
	"fmt"
	"strings"
	"time"

	"captain/apis/cluster/v1alpha1"
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/client/informers/externalversions"
	"captain/pkg/crd"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type clusterProvider struct {
	sharedInformers externalversions.SharedInformerFactory
	client          crd.CrdInterface
}

func New(informer externalversions.SharedInformerFactory, client crd.CrdInterface) clusterProvider {
	return clusterProvider{
		sharedInformers: informer,
		client:          client,
	}
}

func (cp clusterProvider) Get(namespace, name string) (runtime.Object, error) {
	return cp.sharedInformers.Cluster().V1alpha1().Clusters().Lister().Get(name)
}

func (cp clusterProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := cp.sharedInformers.Cluster().V1alpha1().Clusters().Lister().List(query.GetSelector())
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
	cluster, ok := object.(*v1alpha1.Cluster)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(clusterStatus(cluster.Status), string(filter.Value)) == 0
	default:
		return alpha1.DefaultObjectMetaFilter(cluster.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftCluster, ok := left.(*v1alpha1.Cluster)
	if !ok {
		return false
	}
	rightCluster, ok := right.(*v1alpha1.Cluster)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftCluster).After(lastUpdateTime(rightCluster))
	default:
		return alpha1.DefaultObjectMetaCompare(leftCluster.ObjectMeta, rightCluster.ObjectMeta, field)
	}
}

func lastUpdateTime(cluster *v1alpha1.Cluster) time.Time {
	recent := cluster.CreationTimestamp.Time

	for _, condition := range cluster.Status.Conditions {
		if condition.LastUpdateTime.After(recent) {
			recent = condition.LastTransitionTime.Time
		}
	}
	return recent
}

const (
	statusReady    = "ready"
	statusNotReady = "notready"
	statusUnknown  = "unknown"
)

func clusterStatus(status v1alpha1.ClusterStatus) string {
	if len(status.Conditions) == 0 {
		return statusUnknown
	} else if status.Conditions[0].Type == "Ready" && status.Conditions[0].Status == v1.ConditionTrue {
		return statusReady
	} else {
		return statusNotReady
	}
}

func (cp clusterProvider) Create(namespace string, obj runtime.Object) (runtime.Object, error) {
	cluster, err := validation(obj)
	if err != nil {
		return nil, err
	}
	clu, err := cp.client.V1beta1().Clusters().Create(context.TODO(), cluster, metav1.CreateOptions{})
	return clu, err
}

func validation(obj runtime.Object) (*v1alpha1.Cluster, error) {
	cluster, ok := obj.(*v1alpha1.Cluster)
	if !ok {
		return nil, fmt.Errorf("can not get cluster struct from cluster.")
	}

	if len(cluster.Name) == 0 {
		return nil, fmt.Errorf("cluster's name required.")
	}
	if cluster.Labels != nil {
		region := cluster.Labels[v1alpha1.ClusterRegion]
		if len(region) != 0 && !strings.HasPrefix(cluster.Name, region+"-") {
			return nil, fmt.Errorf("error cluster.Name format.(cluster.Name should has prefix be like '{region}-') ")
		}
	}
	// TODO cluster内容校验，连通性校验
	return cluster, nil
}

func (cp clusterProvider) Delete(namespace, name string) error {
	err := cp.client.V1beta1().Clusters().Delete(context.TODO(), name, metav1.DeleteOptions{})
	return err
}

func (cp clusterProvider) Update(namespace, name string, obj runtime.Object) (runtime.Object, error) {

	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	if metaObj.GetName() != name {
		return nil, fmt.Errorf("cluster's name incorrect.")
	}

	cluster, err := validation(obj)
	if err != nil {
		return nil, err
	}

	oldCluster, err := cp.sharedInformers.Cluster().V1alpha1().Clusters().Lister().Get(name)
	if err != nil {
		return nil, err
	}
	cluster.ResourceVersion = oldCluster.ResourceVersion

	clu, err := cp.client.V1beta1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{})
	return clu, err
}
