package v1alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"captain/pkg/constants"
	model "captain/pkg/models/monitoring"
	"captain/pkg/server/runtime"
	"captain/pkg/simple/client/monitoring"
)

const (
	groupName = "monitoring.captain.io"
	respOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: groupName, Version: "v1alpha1"}

func AddToContainer(c *restful.Container, monitoringClient monitoring.Interface) error {
	ws := runtime.NewWebService(GroupVersion)

	h := NewHandler(monitoringClient)

	// cluster
	ws.Route(ws.GET("/cluster").
		To(h.handleClusterMetricsQuery).
		Doc("Get cluster-level metric data.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMetricsTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	// node
	// workload
	// pod
	c.Add(ws)
	return nil
}
