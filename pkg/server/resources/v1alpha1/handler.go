package v1alpha1

import (
	"captain/apis/cluster/v1alpha1"
	"captain/pkg/api"
	"captain/pkg/bussiness/captain-resources/v1alpha1/resource"
	"captain/pkg/unify/query"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

type Handler struct {
	resourceProvider *resource.ResourceProcessor
}

func New(kubeResProcessor *resource.ResourceProcessor) *Handler {
	return &Handler{
		resourceProvider: kubeResProcessor,
	}
}

// handleListResources retrieves resources
func (h *Handler) handleListResources(request *restful.Request, response *restful.Response) {
	query := query.ParseQueryParameter(request)
	resourceType := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")

	result, err := h.resourceProvider.List(resourceType, namespace, query)
	handleResponse(request, response, result, err)
}

func (h *Handler) handleGetResource(request *restful.Request, response *restful.Response) {
	// query := query.ParseQueryParameter(request)
	resourceType := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")

	result, err := h.resourceProvider.Get(resourceType, namespace, name)
	handleResponse(request, response, result, err)
}

func (h *Handler) handleCreateResource(request *restful.Request, response *restful.Response) {
	resource := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")

	obj := getObject(resource)
	if err := request.ReadEntity(obj); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	result, err := h.resourceProvider.Create(resource, namespace, obj)
	handleResponse(request, response, result, err)
}

func (h *Handler) handleDeleteResource(request *restful.Request, response *restful.Response) {
	resource := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")

	err := h.resourceProvider.Delete(resource, namespace, name)
	handleResponse(request, response, nil, err)
}

func (h *Handler) handleUpdateResource(request *restful.Request, response *restful.Response) {
	resource := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")

	obj := getObject(resource)
	if err := request.ReadEntity(obj); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	result, err := h.resourceProvider.Update(resource, namespace, name, obj)
	handleResponse(request, response, result, err)
}

func getObject(resource string) runtime.Object {

	switch resource {
	case v1alpha1.ResourcesPluralCluster:
		return &v1alpha1.Cluster{}
	default:
		return nil
	}
}

func handleResponse(req *restful.Request, resp *restful.Response, obj interface{}, err error) {

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		} else if errors.IsConflict(err) {
			api.HandleConflict(resp, req, err)
			return
		}
		api.HandleBadRequest(resp, req, err)
		return
	}

	_ = resp.WriteEntity(obj)
}
