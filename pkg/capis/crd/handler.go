package crd

import (
	"captain/pkg/api"
	"captain/pkg/controller/crd_controller"
	"captain/pkg/controller/implement/crd"
	"context"
	"k8s.io/client-go/kubernetes"

	"github.com/emicklei/go-restful"
	"k8s.io/klog"
)

type CrdHandler struct {
	crd            crd.Interface
}

func NewCrdHandler(k8sclient kubernetes.Interface) *CrdHandler {
	return &CrdHandler{
		crd:  crd.NewCrdOperator(k8sclient),
	}
}

func (c *CrdHandler) CreateGaiaClusterTemplate(request *restful.Request, response *restful.Response) {
	var gaiaCluster crd_controller.CkeClusterVo

	err := request.ReadEntity(&gaiaCluster)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		cancel()
		return
	}
	created, err := c.crd.CreateGaiaClusterTemplate(ctx, gaiaCluster)
	response.WriteEntity(created)
	cancel()
}
