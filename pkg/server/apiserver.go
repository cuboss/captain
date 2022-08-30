package server

import (
	"context"
	"net/http"
	"strings"

	"captain/pkg/capis/version"
	"captain/pkg/informers"
	captainserverconfig "captain/pkg/server/config"
	"captain/pkg/server/filters"
	"captain/pkg/server/request"
	resAlpha1 "captain/pkg/server/resources/alpha1"
	"captain/pkg/simple/client/k8s"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type CaptainAPIServer struct {
	ServerCount int

	Server *http.Server

	Config *captainserverconfig.Config

	// webservice container, where all webservice defines
	container *restful.Container

	KubernetesClient k8s.Client

	// informerFactory is a collection of all kubernetes(include CRDs) objects informers,
	// mainly for fast query
	InformerFactory informers.CapInformerFactory

	// controller-runtime client
	KubeRuntimeCache cache.Cache
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
	responsewriters.InternalError(w, req, err)
}

func (s *CaptainAPIServer) PrepareRun(stopCh <-chan struct{}) error {
	s.container = restful.NewContainer()
	//s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	//s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
	//	logStackOnRecover(panicReason, httpWriter)
	//})

	// install apis
	s.installCaptainAPIs()

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	s.Server.Handler = s.container

	// handle chain
	s.buildHandlerChain(stopCh)

	return nil
}

// Install all captain api groups
// Installation happens before all informers start to cache objects, so
//   any attempt to list objects using listers will get empty results.
func (s *CaptainAPIServer) installCaptainAPIs() {
	// nataive apis of kubernetes
	urlruntime.Must(version.AddToContainer(s.container, s.KubernetesClient.Discovery()))

	// captain apis for kube resources
	urlruntime.Must(resAlpha1.AddToContainer(s.container, s.InformerFactory, s.KubeRuntimeCache))

}

//通过WithRequestInfo解析API请求的信息，WithKubeAPIServer根据API请求信息判断是否代理请求给Kubernetes
func (s *CaptainAPIServer) buildHandlerChain(stopCh <-chan struct{}) {
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes: sets.NewString("api", "apis"),
	}

	handler := s.Server.Handler
	handler = filters.WithKubeAPIServer(handler, s.KubernetesClient.Config(), &errorResponder{})

	handler = filters.WithRequestInfo(handler, requestInfoResolver)

	s.Server.Handler = handler
}

func (s *CaptainAPIServer) Run(ctx context.Context) (err error) {

	err = s.waitForCacheSync(ctx)

	if err != nil {
		return err
	}

	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = s.Server.Shutdown(shutdownCtx)
	}()

	// Caching resources
	// informersFactory := informers.NewInformerFactories(kubeClient)

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}
	return err
}

func (s *CaptainAPIServer) waitForCacheSync(ctx context.Context) error {

	klog.V(0).Info("starting caching objects")

	stopCh := ctx.Done()
	discoveryCli := s.KubernetesClient.Kubernetes().Discovery()

	_, fullResourcesList, err := discoveryCli.ServerGroupsAndResources()
	if err != nil {
		return err
	}

	isResourceValid := func(gvr schema.GroupVersionResource) bool {
		for _, resource := range fullResourcesList {
			if resource.GroupVersion == gvr.GroupVersion().String() {
				for _, gr := range resource.APIResources {
					if strings.Compare(gr.Name, gvr.Resource) == 0 {
						return true
					}
				}
			}
		}
		return false
	}

	supportedKubeGVRs := []schema.GroupVersionResource{
		{Group: "apps", Version: "v1", Resource: "deployments"},
	}

	//prepare informer for caching
	for _, support := range supportedKubeGVRs {
		if !isResourceValid(support) {
			klog.Warningf("resources %s was not supported in this cluster")
		} else {
			_, err := s.InformerFactory.KubernetesSharedInformerFactory().ForResource(support)
			if err != nil {
				klog.Errorf("can not make informer for resource - %s ", support.String())
			}
		}
	}
	s.InformerFactory.KubernetesSharedInformerFactory().Start(stopCh)

	return nil
}
