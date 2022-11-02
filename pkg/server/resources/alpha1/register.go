package alpha1

import (
	"captain/pkg/api"
	"captain/pkg/bussiness/kube-resources/alpha1/resource"
	"captain/pkg/informers"
	"captain/pkg/server/config"
	"captain/pkg/server/runtime"
	"captain/pkg/unify/query"
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	// "github.com/rogpeppe/go-internal/cache"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	GroupName            = "resources.captain.io"
	ok                   = "success"
	tagClusteredResource = "Resources in cluster scope"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "alpha1"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func AddToContainer(c *restful.Container, factory informers.CapInformerFactory, cache cache.Cache, config *config.Config) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := New(resource.NewResourceProcessor(factory, cache, config))

	webservice.Route(webservice.GET("/namespaces/{namespace}/resources/{resources}").
		To(handler.handleListResources).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("Cluster level resources").
		Param(webservice.PathParameter("resources", "namespace scope resource type, e.g: pods,jobs,configmaps,services.")).
		Param(webservice.PathParameter("namespace", "namespace")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page, which is started with 1 not 0, default value is 1.").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterPageSize, "pageSize").Required(false).DataFormat("pageSize=%d").DefaultValue("pageSize=10")).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, ok, api.ListResult{}))
	webservice.Route(webservice.GET("resources/{resources}").
		To(handler.handleListResources).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("core level resources").
		Param(webservice.PathParameter("resources", "core scope resource type, e.g: namespaces,nodes.")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page, which is started with 1 not 0, default value is 1.").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterPageSize, "pageSize").Required(false).DataFormat("pageSize=%d").DefaultValue("pageSize=10")).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, ok, api.ListResult{}))
	webservice.Route(webservice.GET("/namespaces/{namespace}/resources/{resources}/name/{name}").
		To(handler.handleGetResource).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("Cluster level resources").
		Param(webservice.PathParameter("resources", "namespace scope resource type, e.g: pods,jobs,configmaps,services.")).
		Param(webservice.PathParameter("namespace", "namespace of resources")).
		Param(webservice.PathParameter("name", "name of resources")).
		Returns(http.StatusOK, ok, api.ListResult{}))
	webservice.Route(webservice.GET("resources/{resources}/name/{name}").
		To(handler.handleGetResource).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("Cluster level resources").
		Param(webservice.PathParameter("resources", "core scope resource type, e.g: namespaces,nodes.")).
		Param(webservice.PathParameter("name", "name of resources")).
		Returns(http.StatusOK, ok, api.ListResult{}))

	c.Add(webservice)

	// +region + cluster
	webservice2 := &restful.WebService{}
	webservice2.Path("/regions").Produces(restful.MIME_JSON)
	urlPrefix := "{region}/clusters/{cluster}/capis/" + GroupVersion.String()

	webservice2.Route(webservice2.GET(urlPrefix+"/namespaces/{namespace}/resources/{resources}").
		To(handler.handleListResources).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("Cluster level resources").
		Param(webservice2.PathParameter("region", "region id of cluster")).
		Param(webservice2.PathParameter("cluster", "name of cluster")).
		Param(webservice2.PathParameter("resources", "namespace scope resource type, e.g: pods,jobs,configmaps,services.")).
		Param(webservice2.PathParameter("namespace", "namespace")).
		Param(webservice2.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice2.QueryParameter(query.ParameterPage, "page, which is started with 1 not 0, default value is 1.").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice2.QueryParameter(query.ParameterPageSize, "pageSize").Required(false).DataFormat("pageSize=%d").DefaultValue("pageSize=10")).
		Param(webservice2.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice2.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, ok, api.ListResult{}))
	webservice2.Route(webservice2.GET(urlPrefix+"/resources/{resources}").
		To(handler.handleListResources).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("core level resources").
		Param(webservice2.PathParameter("region", "region id of cluster")).
		Param(webservice2.PathParameter("cluster", "name of cluster")).
		Param(webservice2.PathParameter("resources", "core scope resource type, e.g: namespaces,nodes.")).
		Param(webservice2.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice2.QueryParameter(query.ParameterPage, "page, which is started with 1 not 0, default value is 1.").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice2.QueryParameter(query.ParameterPageSize, "pageSize").Required(false).DataFormat("pageSize=%d").DefaultValue("pageSize=10")).
		Param(webservice2.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice2.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, ok, api.ListResult{}))
	webservice2.Route(webservice2.GET(urlPrefix+"/namespaces/{namespace}/resources/{resources}/name/{name}").
		To(handler.handleGetResource).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("Cluster level resources").
		Param(webservice2.PathParameter("region", "region id of cluster")).
		Param(webservice2.PathParameter("cluster", "name of cluster")).
		Param(webservice2.PathParameter("resources", "namespace scope resource type, e.g: pods,jobs,configmaps,services.")).
		Param(webservice2.PathParameter("namespace", "namespace of resources")).
		Param(webservice2.PathParameter("name", "name of resources")).
		Returns(http.StatusOK, ok, api.ListResult{}))
	webservice2.Route(webservice2.GET(urlPrefix+"/resources/{resources}/name/{name}").
		To(handler.handleGetResource).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("Cluster level resources").
		Param(webservice2.PathParameter("region", "region id of cluster")).
		Param(webservice2.PathParameter("cluster", "name of cluster")).
		Param(webservice2.PathParameter("resources", "core scope resource type, e.g: namespaces,nodes.")).
		Param(webservice2.PathParameter("name", "name of resources")).
		Returns(http.StatusOK, ok, api.ListResult{}))

	c.Add(webservice2)

	return nil
}
