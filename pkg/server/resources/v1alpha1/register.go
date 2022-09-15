package v1alpha1

import (
	"net/http"
	"strings"

	"captain/pkg/api"
	"captain/pkg/bussiness/captain-resources/v1alpha1/resource"
	"captain/pkg/informers"
	"captain/pkg/server/runtime"
	"captain/pkg/simple/client/k8s"
	"captain/pkg/simple/server/errors"
	"captain/pkg/unify/query"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type CaptainResource struct {
	GroupVersion schema.GroupVersion
	Name         string
	Resources    []string
	Namespaced   bool // defult false
}

var resoureces = []CaptainResource{
	{
		GroupVersion: schema.GroupVersion{Group: "cluster.captain.io", Version: "v1alpha1"},
		Name:         "Cluster",
		Resources:    []string{"clusters"},
	},
}

var NamespacedGroupVersions = []schema.GroupVersion{}

func AddToContainer(c *restful.Container, factory informers.CapInformerFactory, client k8s.Client, cache cache.Cache) error {
	handler := New(resource.NewResourceProcessor(factory, client.Crd(), cache))

	for _, resource := range resoureces {
		webservice := runtime.NewWebService(resource.GroupVersion)

		// no namespace scoped
		webservice.Route(webservice.GET("/{resources}").
			To(handler.handleListResources).
			Metadata(restfulspec.KeyOpenAPITags, []string{resource.Name}).
			Doc("list "+strings.Join(resource.Resources, ", ")).
			Param(webservice.PathParameter("resources", "known values include "+strings.Join(resource.Resources, ", "))).
			Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
			Param(webservice.QueryParameter(query.ParameterPage, "page, which is started with 1 not 0, default value is 1.").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
			Param(webservice.QueryParameter(query.ParameterPageSize, "pageSize").Required(false).DataFormat("pageSize=%d").DefaultValue("pageSize=10")).
			Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
			Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
			Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}))

		webservice.Route(webservice.GET("/{resources}/{name}").
			To(handler.handleGetResource).
			Metadata(restfulspec.KeyOpenAPITags, []string{resource.Name}).
			Doc("get single "+strings.Join(resource.Resources, ", ")).
			Param(webservice.PathParameter("resources", "known values include "+strings.Join(resource.Resources, ", "))).
			Param(webservice.PathParameter("name", "name of resources")).
			Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
			Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
			Param(webservice.QueryParameter(query.ParameterPageSize, "pageSize").Required(false).DataFormat("pageSize=%d").DefaultValue("pageSize=10")).
			Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
			Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
			Returns(http.StatusOK, api.StatusOK, nil))

		webservice.Route(webservice.POST("/{resources}").
			To(handler.handleCreateResource).
			Metadata(restfulspec.KeyOpenAPITags, []string{resource.Name}).
			Doc("create "+strings.Join(resource.Resources, ", ")).
			Param(webservice.PathParameter("resources", "known values include "+strings.Join(resource.Resources, ", "))).
			Returns(http.StatusOK, api.StatusOK, nil))

		webservice.Route(webservice.DELETE("/{resources}/{name}").
			To(handler.handleDeleteResource).
			Metadata(restfulspec.KeyOpenAPITags, []string{resource.Name}).
			Doc("delete "+strings.Join(resource.Resources, ", ")).
			Param(webservice.PathParameter("resources", "known values include "+strings.Join(resource.Resources, ", "))).
			Param(webservice.PathParameter(query.ParameterName, "the name of the resource")).
			Returns(http.StatusOK, api.StatusOK, errors.None))

		webservice.Route(webservice.PUT("/{resources}/{name}").
			To(handler.handleUpdateResource).
			Metadata(restfulspec.KeyOpenAPITags, []string{resource.Name}).
			Doc("update "+strings.Join(resource.Resources, ", ")).
			Param(webservice.PathParameter("resources", "known values include "+strings.Join(resource.Resources, ", "))).
			Returns(http.StatusOK, api.StatusOK, nil))

		c.Add(webservice)
	}
	return nil
}
