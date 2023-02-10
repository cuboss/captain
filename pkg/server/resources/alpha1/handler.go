package alpha1

import (
	"captain/pkg/api"
	"captain/pkg/bussiness/kube-resources/alpha1/resource"
	"captain/pkg/unify/query"

	"github.com/emicklei/go-restful"
	"k8s.io/klog"
)

type Handler struct {
	resourceProviderAlpha1 *resource.ResourceProcessor
}

func New(kubeResProcessor *resource.ResourceProcessor) *Handler {
	return &Handler{
		resourceProviderAlpha1: kubeResProcessor,
	}
}

// handleListResources retrieves resources
func (h *Handler) handleListResources(request *restful.Request, response *restful.Response) {
	query := query.ParseQueryParameter(request)
	region := request.PathParameter("region")
	cluster := request.PathParameter("cluster")
	resourceType := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")

	result, err := h.resourceProviderAlpha1.List(region, cluster, resourceType, namespace, query)
	if err == nil {
		response.WriteEntity(result)
		return
	}

	klog.Error(err)
	if err != resource.ErrResourceNotSupported {
		klog.Error(err, resourceType)
		api.HandleInternalError(response, request, err)
		return
	} else {
		api.HandleNotFound(response, request, err)
		return
	}

	// // fallback to v1alpha2
	// result, err = h.fallback(resourceType, namespace, query)
	// if err != nil {
	// 	if err == resourcev1alpha2.ErrResourceNotSupported {
	// 		api.HandleNotFound(response, request, err)
	// 		return
	// 	}
	// 	klog.Error(err)
	// 	api.HandleError(response, request, err)
	// 	return
	// }
	// response.WriteEntity(result)
}

func (h *Handler) handleGetResource(request *restful.Request, response *restful.Response) {
	region := request.PathParameter("region")
	cluster := request.PathParameter("cluster")
	resource := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := h.resourceProviderAlpha1.Get(region, cluster, resource, namespace, name)
	if err != nil {
		response.WriteEntity(result)
		return
	}
	response.WriteEntity(result)
}
