package v1alpha1

import (
	"captain/pkg/api"
	"captain/pkg/capis/component/v1alpha1/tools"
	"captain/pkg/informers"
	model "captain/pkg/models/component"
	"captain/pkg/server/config"
	"captain/pkg/simple/client/helm"
	"captain/pkg/utils/clusterclient"

	"github.com/emicklei/go-restful"
)

type Handler struct {
	clusterclient.ClusterClients
}

func NewHandler(factory informers.CapInformerFactory, config *config.Config) Handler {
	clients := clusterclient.NewClusterClients(factory.CaptainSharedInformerFactory().Cluster().V1alpha1().Clusters(), config.MultiClusterOptions)
	return Handler{
		ClusterClients: clients,
	}
}

func (h Handler) handleClusterComponentInstall(req *restful.Request, resp *restful.Response) {
	regionName := req.PathParameter("region")
	clusterName := req.PathParameter("cluster")
	clusterComponent := &model.ClusterComponent{}
	err := req.ReadEntity(clusterComponent)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
	}

	tools, err := h.NewComponentTool(regionName, clusterName, clusterComponent)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	release, err := tools.Install()
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	resp.WriteEntity(release)
}

func (h Handler) handleClusterComponentUpgrade(req *restful.Request, resp *restful.Response) {
	// TODO upgrade
}

// 卸载组件
func (h Handler) handleClusterComponentUninstall(req *restful.Request, resp *restful.Response) {
	regionName := req.PathParameter("region")
	clusterName := req.PathParameter("cluster")
	clusterComponent := &model.ClusterComponent{}
	err := req.ReadEntity(clusterComponent)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
	}

	tools, err := h.NewComponentTool(regionName, clusterName, clusterComponent)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	release, err := tools.Uninstall()
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	resp.WriteEntity(release)
}

func (h Handler) handleClusterComponentStatus(req *restful.Request, resp *restful.Response) {
	regionName := req.PathParameter("region")
	clusterName := req.PathParameter("cluster")
	releaseName := req.PathParameter("release")
	clusterComponent := &model.ClusterComponent{}
	err := req.ReadEntity(clusterComponent)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	tools, err := h.NewComponentTool(regionName, clusterName, clusterComponent)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	res, err := tools.Status(releaseName)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	resp.WriteEntity(model.ClusterComponentResources{Resources: res})
}

func (h Handler) NewComponentTool(regionName, clusterName string, clusterComponent *model.ClusterComponent) (tools.Interface, error) {
	// init client
	cluster, err := h.Get(regionName, clusterName)
	if err != nil {
		return nil, err
	}
	kubeConfig := cluster.Spec.Connection.KubeConfig
	client, err := helm.NewClient(kubeConfig, clusterComponent.Namespace)
	if err != nil {
		return nil, err
	}

	switch clusterComponent.ComponentName {
	case "prometheus":
		return tools.NewPrometheus(client, clusterComponent)

		// TOTO ADD MORE Component
	}

	return nil, nil
}
