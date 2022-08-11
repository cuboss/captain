package crd

import (
	"captain/apis/gaia/v1alpha1"
	"captain/pkg/controller/crd_controller"
	"captain/pkg/server/runtime"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

const (
	GroupName = "gaia.welkin"
	Version   = "v1alpha1"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func AddToContainer(container *restful.Container, k8sclient kubernetes.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	crdHandler := NewCrdHandler(k8sclient)

	webservice.Route(webservice.POST("/gaiaclusters").
		To(crdHandler.CreateGaiaClusterTemplate).
		Doc("Create gaiaclusters.").
		Reads(crd_controller.CkeClusterVo{}).
		Returns(http.StatusOK, "ok", v1alpha1.GaiaCluster{}).
		Metadata(restfulspec.KeyOpenAPITags, "create gaiacluster"))

	container.Add(webservice)

	return nil
}