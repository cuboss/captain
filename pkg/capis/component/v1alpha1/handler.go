package v1alpha1

import (
	"captain/pkg/informers"
	model "captain/pkg/models/component"
	"captain/pkg/server/config"
	"captain/pkg/simple/client/helm"
	"captain/pkg/utils/clusterclient"
	"encoding/json"

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
		// TODO return error
	}
	// init client
	cluster, err := h.Get(regionName, clusterName)
	if err != nil {
		// TODO return error
	}
	kubeConfig := cluster.Spec.Connection.KubeConfig
	clien, err := helm.NewClient(kubeConfig)
	if err != nil {
		// TODO return error
	}
	// install
	values := make(map[string]interface{})
	err = json.Unmarshal([]byte(clusterComponent.Values), &values)
	if err != nil {
		// TODO return error
	}
	release, err := clien.Install(clusterComponent.ReleaseName, clusterComponent.ComponentName, clusterComponent.ComponentVersion, clusterComponent.Namespace, values)
	resp.WriteEntity(release)
}

func (h Handler) handleClusterComponentUpgrade(req *restful.Request, resp *restful.Response) {
	// TODO upgrade
}

func (h Handler) handleClusterComponentUninstall(req *restful.Request, resp *restful.Response) {
	// TODO install
}
func (h Handler) handleClusterComponentStatus(req *restful.Request, resp *restful.Response) {
	// TODO fetch status
}
