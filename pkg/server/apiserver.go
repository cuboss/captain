package server

import (
	"context"
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"

	"captain/pkg/capis/version"
	captainserverconfig "captain/pkg/server/config"
	"captain/pkg/server/dispatch"
	"captain/pkg/server/filters"
	"captain/pkg/server/informers"
	"captain/pkg/server/request"
	"captain/pkg/simple/client/k8s"
)

type APIServer struct {
	ServerCount int

	Server *http.Server

	Config *captainserverconfig.Config

	// webservice container, where all webservice defines
	container *restful.Container

	// kubeClient is a collection of all kubernetes(include CRDs) objects clientset
	KubernetesClient k8s.Client

	// informerFactory is a collection of all kubernetes(include CRDs) objects informers,
	// mainly for fast query
	InformerFactory informers.InformerFactory
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
	responsewriters.InternalError(w, req, err)
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {
	s.container = restful.NewContainer()
	//s.container.Filter(logRequestAndResponse)
	// 设定路由为CurlyRouter(快速路由)
	s.container.Router(restful.CurlyRouter{})
	//s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
	//	logStackOnRecover(panicReason, httpWriter)
	//})

	//s.installKubeSphereAPIs(stopCh)
	//s.installCRDAPIs()
	//s.installMetricsAPI()
	//s.container.Filter(monitorRequest)

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	// container 作为http server 的handler
	s.Server.Handler = s.container

	// 注册服务
	s.installCaptainAPIs()
	// handle chain
	s.buildHandlerChain(stopCh)

	return nil
}

// Install all captain api groups
// Installation happens before all informers start to cache objects, so
//   any attempt to list objects using listers will get empty results.
func (s *APIServer) installCaptainAPIs() {
	urlruntime.Must(version.AddToContainer(s.container, s.KubernetesClient.Discovery()))
}

//通过WithRequestInfo解析API请求的信息，WithKubeAPIServer根据API请求信息判断是否代理请求给Kubernetes
func (s *APIServer) buildHandlerChain(stopCh <-chan struct{}) {
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes: sets.NewString("api", "apis"),
	}

	handler := s.Server.Handler
	handler = filters.WithKubeAPIServer(handler, s.KubernetesClient.Config(), &errorResponder{})

	if s.Config.MultiClusterOptions.Enable {
		clusterDispatcher := dispatch.NewClusterDispatch(s.InformerFactory.CaptainSharedInformerFactory().Cluster().V1alpha1().Clusters())
		handler = filters.WithMultipleClusterDispatcher(handler, clusterDispatcher)
	}

	handler = filters.WithRequestInfo(handler, requestInfoResolver)

	s.Server.Handler = handler
}

func (s *APIServer) waitForResourceSync(ctx context.Context) error {
	klog.V(0).Info("Start cache objects")

	stopCh := ctx.Done()

	discoveryClient := s.KubernetesClient.Kubernetes().Discovery()
	_, apiResourcesList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return err
	}

	isResourceExists := func(resource schema.GroupVersionResource) bool {
		for _, apiResource := range apiResourcesList {
			if apiResource.GroupVersion == resource.GroupVersion().String() {
				for _, rsc := range apiResource.APIResources {
					if rsc.Name == resource.Resource {
						return true
					}
				}
			}
		}
		return false
	}

	crdInformerFactory := s.InformerFactory.CaptainSharedInformerFactory()

	captainGVRs := []schema.GroupVersionResource{
		{Group: "cluster.captain.io", Version: "v1beta1", Resource: "clusters"},
	}

	for _, gvr := range captainGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err = crdInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}

	crdInformerFactory.Start(stopCh)
	crdInformerFactory.WaitForCacheSync(stopCh)

	klog.V(0).Info("Finished caching objects")

	return nil
}

func (s *APIServer) Run(ctx context.Context) (err error) {
	err = s.waitForResourceSync(ctx)
	if err != nil {
		return err
	}

	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = s.Server.Shutdown(shutdownCtx)
	}()

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}

	return err
}
