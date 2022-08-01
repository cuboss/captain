package kubeconfig

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Interface interface {
	GetKubeConfig(username string) (string, error)
	CreateKubeConfig(user *iamv1alpha2.User) error
	UpdateKubeconfig(username string, csr *certificatesv1.CertificateSigningRequest) error
}

type operator struct {
	k8sClient       kubernetes.Interface
	configMapLister corev1listers.ConfigMapLister
	config          *rest.Config
	masterURL       string
}

func NewOperator(k8sClient kubernetes.Interface, configMapLister corev1listers.ConfigMapLister, config *rest.Config) Interface {
	return &operator{k8sClient: k8sClient, configMapLister: configMapLister, config: config}
}
