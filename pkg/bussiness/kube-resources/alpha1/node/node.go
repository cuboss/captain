package node

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type nodeProvider struct {
	informers informers.SharedInformerFactory
}

const (
	nodeConfigOK        v1.NodeConditionType = "ConfigOK"
	nodeKubeletReady    v1.NodeConditionType = "KubeletReady"
	StatusUnschedulable                      = "unschedulable"
	StatusWarning                            = "warning"
	StatusRunning                            = "running"
)

func New(informer informers.SharedInformerFactory) nodeProvider {
	return nodeProvider{informers: informer}
}

func (nd nodeProvider) Get(_, name string) (runtime.Object, error) {
	return nd.informers.Core().V1().Nodes().Lister().Get(name)

}

func (nd nodeProvider) List(node string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := nd.informers.Core().V1().Nodes().Lister().List(query.GetSelector())
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
	node, ok := object.(*v1.Node)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldRole:
		labelKey := fmt.Sprintf("node-role.kubernetes.io/%s", filter.Value)
		if _, ok := node.Labels[labelKey]; !ok {
			return false
		} else {
			return true
		}
	case query.FieldStatus:
		return strings.Compare(getNodeStatus(node), string(filter.Value)) == 0
	default:
		return alpha1.DefaultObjectMetaFilter(node.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftND, ok := left.(*v1.Node)
	if !ok {
		return false
	}
	rightND, ok := right.(*v1.Node)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftND).After(lastUpdateTime(rightND))
	default:
		return alpha1.DefaultObjectMetaCompare(leftND.ObjectMeta, rightND.ObjectMeta, field)
	}
}

func lastUpdateTime(node *v1.Node) time.Time {
	recent := node.CreationTimestamp.Time

	for _, condition := range node.Status.Conditions {
		if condition.LastTransitionTime.After(recent) {
			recent = condition.LastTransitionTime.Time
		}
	}
	return recent
}

var expectedConditions = map[v1.NodeConditionType]v1.ConditionStatus{
	v1.NodeMemoryPressure:     v1.ConditionFalse,
	v1.NodeDiskPressure:       v1.ConditionFalse,
	v1.NodePIDPressure:        v1.ConditionFalse,
	v1.NodeNetworkUnavailable: v1.ConditionFalse,
	nodeConfigOK:              v1.ConditionTrue,
	nodeKubeletReady:          v1.ConditionTrue,
	v1.NodeReady:              v1.ConditionTrue,
}

func isUnhealthyStatus(condition v1.NodeCondition) bool {
	expectedStatus := expectedConditions[condition.Type]
	if expectedStatus != "" && condition.Status != expectedStatus {
		return true
	}
	return false
}

func getNodeStatus(node *v1.Node) string {
	if node.Spec.Unschedulable {
		return StatusUnschedulable
	}
	for _, condition := range node.Status.Conditions {
		if isUnhealthyStatus(condition) {
			return StatusWarning
		}
	}

	return StatusRunning
}
