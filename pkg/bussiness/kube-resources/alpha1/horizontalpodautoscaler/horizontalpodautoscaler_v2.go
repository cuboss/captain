package horizontalpodautoscaler

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type hpaV2Provider struct {
	sharedInformers informers.SharedInformerFactory
}

func NewV2HpaProvider(informer informers.SharedInformerFactory) hpaV2Provider {
	return hpaV2Provider{sharedInformers: informer}
}

func (hpa hpaV2Provider) Get(namespace, name string) (runtime.Object, error) {
	return hpa.sharedInformers.Autoscaling().V2().HorizontalPodAutoscalers().Lister().HorizontalPodAutoscalers(namespace).Get(name)
}

func (hpa hpaV2Provider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := hpa.sharedInformers.Autoscaling().V2().HorizontalPodAutoscalers().Lister().HorizontalPodAutoscalers(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, deploy := range raw {
		result = append(result, deploy)
	}

	return alpha1.DefaultList(result, query, v2CompareFunc, v2Filter), nil
}

func v2Filter(object runtime.Object, filter query.Filter) bool {
	ingress, ok := object.(*v2.HorizontalPodAutoscaler)
	if !ok {
		return false
	}
	return alpha1.DefaultObjectMetaFilter(ingress.ObjectMeta, filter)
}

func v2CompareFunc(left, right runtime.Object, field query.Field) bool {

	leftHpa, ok := left.(*v2.HorizontalPodAutoscaler)
	if !ok {
		return false
	}

	rightHpa, ok := right.(*v2.HorizontalPodAutoscaler)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	default:
		return alpha1.DefaultObjectMetaCompare(leftHpa.ObjectMeta, rightHpa.ObjectMeta, field)
	}
}
