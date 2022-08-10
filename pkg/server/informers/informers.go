package informers

import (
	"time"

	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	"captain/pkg/client/informers/externalversions"
	"captain/pkg/crd"
)

// default re-sync period for all informer factories
const defaultResync = 600 * time.Second

// InformerFactory is a group all shared informer factories which captain needed
// callers should check if the return value is nil
type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	CaptainSharedInformerFactory() externalversions.SharedInformerFactory

	// Start shared informer factory one by one if they are not nil
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	informerFactory k8sinformers.SharedInformerFactory
	captainFactory  externalversions.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, crdClient crd.CrdInterface) InformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}

	if crdClient != nil {
		factory.captainFactory = externalversions.NewSharedInformerFactory(crdClient.Versioned(), defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.informerFactory
}

func (f *informerFactories) CaptainSharedInformerFactory() externalversions.SharedInformerFactory {
	return f.captainFactory
}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.informerFactory != nil {
		f.informerFactory.Start(stopCh)
	}

	if f.captainFactory != nil {
		f.captainFactory.Start(stopCh)
	}
}
