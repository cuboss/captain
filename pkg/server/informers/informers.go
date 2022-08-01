package informers

import (
	"time"

	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	"captain/pkg/client/clientset/versioned"
	captaininformers "captain/pkg/client/informers/externalversions"
)

// default re-sync period for all informer factories
const defaultResync = 600 * time.Second

// InformerFactory is a group all shared informer factories which captain needed
// callers should check if the return value is nil
type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	CaptainSharedInformerFactory() captaininformers.SharedInformerFactory

	// Start shared informer factory one by one if they are not nil
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	informerFactory        k8sinformers.SharedInformerFactory
	captainInformerFactory captaininformers.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, captainClient versioned.Interface) InformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}

	if captainClient != nil {
		factory.captainInformerFactory = captaininformers.NewSharedInformerFactory(captainClient, defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.informerFactory
}

func (f *informerFactories) CaptainSharedInformerFactory() captaininformers.SharedInformerFactory {
	return f.captainInformerFactory
}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.informerFactory != nil {
		f.informerFactory.Start(stopCh)
	}

	if f.captainInformerFactory != nil {
		f.captainInformerFactory.Start(stopCh)
	}
}
