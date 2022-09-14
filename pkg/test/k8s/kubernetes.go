package k8s

import (
	"captain/pkg/crd"
	"captain/pkg/simple/client/k8s"

	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type testKubernetesClient struct {
	// kubernetes client interface
	k8s kubernetes.Interface

	// kubernetes crd interface
	crd crd.CrdInterface

	// discovery client
	discoveryClient *discovery.DiscoveryClient

	istio istioclient.Interface

	snapshot snapshotclient.Interface

	apiextensions apiextensionsclient.Interface

	prometheus promresourcesclient.Interface

	master string

	config *rest.Config
}

// NewKubernetesClient creates a KubernetesClient
func NewTestKubernetesClient(cli kubernetes.Interface) (k8s.Client, error) {
	var k testKubernetesClient

	k.k8s = cli
	k.crd = 
}

func (k *testKubernetesClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *testKubernetesClient) Crd() crd.CrdInterface {
	return k.crd
}

func (k *testKubernetesClient) Discovery() discovery.DiscoveryInterface {
	return k.discoveryClient
}

func (k *testKubernetesClient) Istio() istioclient.Interface {
	return k.istio
}

func (k *testKubernetesClient) Snapshot() snapshotclient.Interface {
	return k.snapshot
}

func (k *testKubernetesClient) ApiExtensions() apiextensionsclient.Interface {
	return k.apiextensions
}

func (k *testKubernetesClient) Prometheus() promresourcesclient.Interface {
	return k.prometheus
}

// master address used to generate kubeconfig for downloading
func (k *testKubernetesClient) Master() string {
	return k.master
}

func (k *testKubernetesClient) Config() *rest.Config {
	return k.config
}
