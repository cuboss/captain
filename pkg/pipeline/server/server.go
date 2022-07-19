package server

import (
	"fmt"
	"log"
	"net/http"

	// "github.com/docker/cli/kubernetes/client/informers"
	"github.com/emicklei/go-restful"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type CapServer struct {
	Server       *http.Server
	KubeInformer *informers.SharedInformerFactory
	KubeClient   kubernetes.Clientset
	// webservice container, where all webservice defines
	container *restful.Container
}

func NewServer(port int, kubeconfig string) (*CapServer, error) {
	conf, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	server := CapServer{}

	kubeClient, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}
	server.KubeClient = *kubeClient

	server.Server = &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	return &server, nil
}

func (s *CapServer) initHandler() http.Handler {
	container := restful.NewContainer()
	// Todo: add default core resources of kubernetes

	// Todo: add restful webservice
	log.Printf("web containers info: %v", container)

	// Todo: handler contructor
	return nil
}

func (s *CapServer) registKubeAPIs() error {
	// Todo: register core resources api handler for CapServer

	return nil

}
