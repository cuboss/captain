

package filters

import (
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"captain/pkg/server/request"
	"captain/pkg/simple/server/errors"
)

// WithKubeAPIServer proxy request to kubernetes service if requests path starts with /api
func WithKubeAPIServer(handler http.Handler, config *rest.Config, failed proxy.ErrorResponder) http.Handler {
	kubernetes, _ := url.Parse(config.Host)
	defaultTransport, err := rest.TransportFor(config)
	if err != nil {
		klog.Errorf("Unable to create transport from rest.Config: %v", err)
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		info, ok := request.RequestInfoFrom(req.Context())
		if !ok {
			err := errors.New("Unable to retrieve request info from request")
			klog.Error(err)
			responsewriters.InternalError(w, req, err)
		}

		if info.IsKubernetesRequest {
			s := *req.URL
			s.Host = kubernetes.Host
			s.Scheme = kubernetes.Scheme

			// make sure we don't override kubernetes's authorization
			req.Header.Del("Authorization")
			httpProxy := proxy.NewUpgradeAwareHandler(&s, defaultTransport, true, false, failed)
			httpProxy.UpgradeTransport = proxy.NewUpgradeRequestRoundTripper(defaultTransport, defaultTransport)
			httpProxy.ServeHTTP(w, req)
			return
		}

		handler.ServeHTTP(w, req)
	})
}
