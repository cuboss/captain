package filters

import (
	"fmt"
	"net/http"

	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"

	"captain/pkg/server/dispatch"
	"captain/pkg/server/request"
	"captain/pkg/server/runtime"
)

// Multiple cluster dispatcher forward request to desired cluster based on request cluster name
// which included in request path clusters/{cluster}
func WithMultipleClusterDispatcher(handler http.Handler, dispatch dispatch.Dispatcher) http.Handler {
	if dispatch == nil {
		klog.V(4).Infof("Multiple cluster dispatcher is disabled")
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		info, ok := request.RequestInfoFrom(req.Context())
		if !ok {
			responsewriters.InternalError(w, req, fmt.Errorf(""))
			return
		}

		if info.Cluster == "" || len(info.APIPrefix) == 0 || (info.APIPrefix == runtime.ApiRoot && !proxyAPIGroup(info.APIGroup)) {
			handler.ServeHTTP(w, req)
		} else {
			dispatch.Dispatch(w, req, handler)
		}
	})
}

var proxyAPIGroups = map[string]bool{
	"monitoring.captain.io": true,
}

// .../capis/monitoring.captain.io/v1alpha1/...
func proxyAPIGroup(group string) bool {
	return proxyAPIGroups[group]
}
