package server

import (
	"context"
	"net/http"

	"github.com/emicklei/go-restful"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"

	"captain/pkg/capis/version"
	captainserverconfig "captain/pkg/server/config"
	"captain/pkg/server/filters"
	"captain/pkg/server/request"
	"captain/pkg/simple/client/k8s"
)

type CaptainAPIServer struct {
	ServerCount int

	Server *http.Server

	Config *captainserverconfig.Config

	// webservice container, where all webservice defines
	container *restful.Container

	KubernetesClient k8s.Client
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
	responsewriters.InternalError(w, req, err)
}

func (s *CaptainAPIServer) PrepareRun(stopCh <-chan struct{}) error {
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
func (s *CaptainAPIServer) installCaptainAPIs() {
	urlruntime.Must(version.AddToContainer(s.container, s.KubernetesClient.Discovery()))
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
