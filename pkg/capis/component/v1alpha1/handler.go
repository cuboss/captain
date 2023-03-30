package v1alpha1

import (
	clusterv1alpha1 "captain/apis/cluster/v1alpha1"
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
	// init client
	cluster, err := h.Get(regionName, clusterName)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	tools, err := NewComponentTool(cluster, clusterComponent)
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

func (h Handler) handleClusterComponentUninstall(req *restful.Request, resp *restful.Response) {
	// TODO install
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
	// init client
	cluster, err := h.Get(regionName, clusterName)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	tools, err := NewComponentTool(cluster, clusterComponent)
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

func NewComponentTool(cluster *clusterv1alpha1.Cluster, clusterComponent *model.ClusterComponent) (tools.Interface, error) {
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
