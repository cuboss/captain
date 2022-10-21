package server

import (
	"context"
	"net/http"

	"captain/pkg/capis/version"
	"captain/pkg/informers"
	captainserverconfig "captain/pkg/server/config"
	"captain/pkg/server/dispatch"
	"captain/pkg/server/filters"
	"captain/pkg/server/request"
	resAlpha1 "captain/pkg/server/resources/alpha1"
	resV1alpha1 "captain/pkg/server/resources/v1alpha1"
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

	// kubeClient is a collection of all kubernetes(include CRDs) objects clientset
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
//
//	any attempt to list objects using listers will get empty results.
func (s *CaptainAPIServer) installCaptainAPIs() {
	// nataive apis of kubernetes
	urlruntime.Must(version.AddToContainer(s.container, s.KubernetesClient.Discovery()))

	// captain apis for kube resources
	urlruntime.Must(resAlpha1.AddToContainer(s.container, s.InformerFactory, s.KubeRuntimeCache))

	// captain apis for captain cluster resources
	urlruntime.Must(resV1alpha1.AddToContainer(s.container, s.InformerFactory, s.KubernetesClient, s.KubeRuntimeCache))

}

// 通过WithRequestInfo解析API请求的信息，WithKubeAPIServer根据API请求信息判断是否代理请求给Kubernetes
func (s *CaptainAPIServer) buildHandlerChain(stopCh <-chan struct{}) {
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

func (s *CaptainAPIServer) waitForResourceSync(ctx context.Context) error {
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

	//caching kubernetes native resources
	kubeGVRs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "namespaces"},
		{Group: "", Version: "v1", Resource: "nodes"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
		{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
		{Group: "", Version: "v1", Resource: "persistentvolumes"},

		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "replicasets"},
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "", Version: "v1", Resource: "serviceaccounts"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	}
	for _, gvr := range kubeGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr.String())
		} else {
			_, err = s.InformerFactory.KubernetesSharedInformerFactory().ForResource(gvr)
			if err != nil {
				klog.Errorf("can not make informer for resource - %s ", gvr.String())
			}
		}
	}
	s.InformerFactory.KubernetesSharedInformerFactory().Start(stopCh)

	// caching other crds
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

func (s *CaptainAPIServer) Run(ctx context.Context) (err error) {
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
