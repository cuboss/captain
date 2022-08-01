package k8s

import (
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	captain "captain/pkg/client/clientset/versioned"
	"captain/pkg/crd"
)

type Client interface {
	Kubernetes() kubernetes.Interface
	Crd() crd.CrdInterface
	Captain() captain.Interface
	Istio() istioclient.Interface
	Snapshot() snapshotclient.Interface
	ApiExtensions() apiextensionsclient.Interface
	Discovery() discovery.DiscoveryInterface
	Prometheus() promresourcesclient.Interface
	Master() string
	Config() *rest.Config
}

type kubernetesClient struct {
	// kubernetes client interface
	k8s kubernetes.Interface

	// kubernetes crd interface
	crd crd.CrdInterface

	// discovery client
	discoveryClient *discovery.DiscoveryClient

	// generated clientset
	captain captain.Interface

	istio istioclient.Interface

	snapshot snapshotclient.Interface

	apiextensions apiextensionsclient.Interface

	prometheus promresourcesclient.Interface

	master string

	config *rest.Config
}

// NewKubernetesClient creates a KubernetesClient
func NewKubernetesClient(options *KubernetesOptions) (Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		return nil, err
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	var k kubernetesClient
	k.k8s, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.discoveryClient, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	k.captain, err = captain.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.istio, err = istioclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	crdClients, err := crd.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	k.crd = crdClients

	k.snapshot, err = snapshotclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.apiextensions, err = apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.prometheus, err = promresourcesclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.master = options.Master
	k.config = config

	return &k, nil
}

func (k *kubernetesClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *kubernetesClient) Crd() crd.CrdInterface {
	return k.crd
}

func (k *kubernetesClient) Discovery() discovery.DiscoveryInterface {
	return k.discoveryClient
}

func (k *kubernetesClient) Captain() captain.Interface {
	return k.captain
}

func (k *kubernetesClient) Istio() istioclient.Interface {
	return k.istio
}

func (k *kubernetesClient) Snapshot() snapshotclient.Interface {
	return k.snapshot
}

func (k *kubernetesClient) ApiExtensions() apiextensionsclient.Interface {
	return k.apiextensions
}

func (k *kubernetesClient) Prometheus() promresourcesclient.Interface {
	return k.prometheus
}

// master address used to generate kubeconfig for downloading
func (k *kubernetesClient) Master() string {
	return k.master
}

func (k *kubernetesClient) Config() *rest.Config {
	return k.config
}
