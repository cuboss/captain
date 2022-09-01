package deployment

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"strings"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type deployProvider struct {
	sharedInformers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) deployProvider {
	return deployProvider{sharedInformers: informer}
}

func (dp deployProvider) Get(namespace, name string) (runtime.Object, error) {
	return dp.sharedInformers.Apps().V1().Deployments().Lister().Deployments(namespace).Get(name)
}

func (dp deployProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := dp.sharedInformers.Apps().V1().Deployments().Lister().Deployments(namespace).List(query.GetSelector())
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
	deployment, ok := object.(*v1.Deployment)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(deploymentStatus(deployment.Status), string(filter.Value)) == 0
	default:
		return alpha1.DefaultObjectMetaFilter(deployment.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftDeploy, ok := left.(*v1.Deployment)
	if !ok {
		return false
	}
	rightDeploy, ok := right.(*v1.Deployment)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftDeploy).After(lastUpdateTime(rightDeploy))
	default:
		return alpha1.DefaultObjectMetaCompare(leftDeploy.ObjectMeta, rightDeploy.ObjectMeta, field)
	}
}

func lastUpdateTime(deploy *v1.Deployment) time.Time {
	recent := deploy.CreationTimestamp.Time

	for _, condition := range deploy.Status.Conditions {
		if condition.LastUpdateTime.After(recent) {
			recent = condition.LastTransitionTime.Time
		}
	}
	return recent
}

func deploymentStatus(status v1.DeploymentStatus) string {
	if status.ReadyReplicas == 0 && status.Replicas == 0 {
		return statusStopped
	} else if status.ReadyReplicas == status.Replicas {
		return statusRunning
	} else {
		return statusUpdating
	}
}
