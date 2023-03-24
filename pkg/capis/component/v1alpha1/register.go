package v1alpha1

import (
	"net/http"

	"captain/pkg/constants"
	"captain/pkg/informers"
	model "captain/pkg/models/component"
	"captain/pkg/server/config"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	groupName = "clustercomponent.captain.io"
	respOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: groupName, Version: "v1alpha1"}

func AddToContainer(c *restful.Container, factory informers.CapInformerFactory, config *config.Config) {
	h := NewHandler(factory, config)

	ws := &restful.WebService{}
	ws.Path("/regions/{region}/clusters/{cluster}/capis/" + GroupVersion.String()).
		Param(ws.PathParameter("region", "region id of cluster")).
		Param(ws.PathParameter("cluster", "name of cluster")).
		Produces(restful.MIME_JSON)

	// 安装
	ws.Route(ws.POST("/clustercomponents").
		To(h.handleClusterComponentInstall).
		Doc("Install component in cluster.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMetricsTag}).
		Writes(model.ClusterComponent{}).
		Returns(http.StatusOK, respOK, model.ClusterComponent{})).
		Produces(restful.MIME_JSON)
	// 升级
	ws.Route(ws.PUT("/clustercomponents/{releaseName}").
		To(h.handleClusterComponentUpgrade).
		Doc("Install component in cluster.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMetricsTag}).
		Writes(model.ClusterComponent{}).
		Returns(http.StatusOK, respOK, model.ClusterComponent{})).
		Produces(restful.MIME_JSON)
	// 卸载
	ws.Route(ws.DELETE("/clustercomponents/{releaseName}").
		To(h.handleClusterComponentUninstall).
		Doc("Install component in cluster.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMetricsTag}).
		Writes(model.ClusterComponent{}).
		Returns(http.StatusOK, respOK, model.ClusterComponent{})).
		Produces(restful.MIME_JSON)
	// 查询
	ws.Route(ws.GET("/clustercomponents/{releaseName}").
		To(h.handleClusterComponentStatus).
		Doc("Install component in cluster.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMetricsTag}).
		Writes(model.ClusterComponent{}).
		Returns(http.StatusOK, respOK, model.ClusterComponent{})).
		Produces(restful.MIME_JSON)
}
