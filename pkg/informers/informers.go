package informers

import (
	"time"

	"captain/pkg/client/informers/externalversions"

	"captain/pkg/crd"

	kubeInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

// default re-sync period for all informer factories
const defaultResync = 5 * time.Second

type CapInformerFactory interface {
	KubernetesSharedInformerFactory() kubeInformers.SharedInformerFactory
	CaptainSharedInformerFactory() externalversions.SharedInformerFactory

	// Start shared informer factory one by one if they are not nil
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	informerFactory kubeInformers.SharedInformerFactory
	captainFactory  externalversions.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, crdClient crd.CrdInterface) CapInformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = kubeInformers.NewSharedInformerFactory(client, defaultResync)
	}

	if crdClient != nil {
		factory.captainFactory = externalversions.NewSharedInformerFactory(crdClient.Versioned(), defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() kubeInformers.SharedInformerFactory {
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
