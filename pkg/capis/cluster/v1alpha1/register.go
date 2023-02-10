package v1alpha1

import (
	"captain/pkg/api"
	"captain/pkg/informers"
	"captain/pkg/server/config"
	"captain/pkg/server/runtime"
	"captain/pkg/utils/clusterclient"
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	GroupName = "cluster.captain.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func AddToContainer(container *restful.Container, factory informers.CapInformerFactory, config *config.Config) {
	clients := clusterclient.NewClusterClients(factory.CaptainSharedInformerFactory().Cluster().V1alpha1().Clusters(), config.MultiClusterOptions)
	h := NewHandler(clients)

	webservice := &restful.WebService{}
	webservice.Path(runtime.ApiRootPath + "/" + GroupVersion.String() + "/clusters").
		Produces(restful.MIME_JSON)

	webservice.Route(webservice.GET("/{name}/adminToken").
		To(h.clusterAdminToken).
		Metadata(restfulspec.KeyOpenAPITags, []string{"clusters"}).
		Doc("get cluster admin token").
		Param(webservice.PathParameter("name", "name of cluster")).
		Param(webservice.QueryParameter("dryRun", "dry run request or not").Required(false)).
		Returns(http.StatusOK, api.StatusOK, nil))

	container.Add(webservice)
}
