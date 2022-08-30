package real

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	//default: 5
	KubeQps = 10
	//default: 10
	KubeBurst = 20

	KubeConfigPath = "~/.kube/config"
)

func NewClientSet() (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", KubeConfigPath)
	if err != nil {
		return nil, err
	}

	config.QPS = float32(KubeQps)
	config.Burst = KubeBurst

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
