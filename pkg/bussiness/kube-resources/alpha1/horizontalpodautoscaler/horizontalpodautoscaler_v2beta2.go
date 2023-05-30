package horizontalpodautoscaler

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"k8s.io/api/autoscaling/v2beta2"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type hpaV2beta2Provider struct {
	sharedInformers informers.SharedInformerFactory
}

func NewV2beta2HpaProvider(informer informers.SharedInformerFactory) hpaV2beta2Provider {
	return hpaV2beta2Provider{sharedInformers: informer}
}

func (hpa hpaV2beta2Provider) Get(namespace, name string) (runtime.Object, error) {
	return hpa.sharedInformers.Autoscaling().V2beta2().HorizontalPodAutoscalers().Lister().HorizontalPodAutoscalers(namespace).Get(name)
}

func (hpa hpaV2beta2Provider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	raw, err := hpa.sharedInformers.Autoscaling().V2beta2().HorizontalPodAutoscalers().Lister().HorizontalPodAutoscalers(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, deploy := range raw {
		result = append(result, deploy)
	}

	return alpha1.DefaultList(result, query, v2beta2CompareFunc, v2beta2Filter), nil
}

func v2beta2Filter(object runtime.Object, filter query.Filter) bool {
	ingress, ok := object.(*v2beta2.HorizontalPodAutoscaler)
	if !ok {
		return false
	}
	return alpha1.DefaultObjectMetaFilter(ingress.ObjectMeta, filter)
}

func v2beta2CompareFunc(left, right runtime.Object, field query.Field) bool {

	leftHpa, ok := left.(*v2beta2.HorizontalPodAutoscaler)
	if !ok {
		return false
	}

	rightHpa, ok := right.(*v2beta2.HorizontalPodAutoscaler)
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
