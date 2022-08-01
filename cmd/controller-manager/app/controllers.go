package app

import (
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"captain/pkg/controller/cluster"
	"captain/pkg/server/informers"
	"captain/pkg/simple/client/k8s"
	"captain/pkg/simple/client/multicluster"
)

func addControllers(
	mgr manager.Manager,
	client k8s.Client,
	informerFactory informers.InformerFactory,
	options *k8s.KubernetesOptions,
	multiClusterOptions *multicluster.Options,
	stopCh <-chan struct{}) error {

	captainInformer := informerFactory.CaptainSharedInformerFactory()

	multiClusterEnabled := multiClusterOptions.Enable

	var clusterController manager.Runnable
	if multiClusterEnabled {
		clusterController = cluster.NewClusterController(
			client.Kubernetes(),
			client.Config(),
			captainInformer.Cluster().V1alpha1().Clusters(),
			client.Captain().ClusterV1alpha1().Clusters(),
			multiClusterOptions.ClusterControllerResyncPeriod,
			multiClusterOptions.HostClusterName)
	}

	controllers := map[string]manager.Runnable{
		"cluster-controller": clusterController,
	}

	for name, ctrl := range controllers {
		if ctrl == nil {
			klog.V(4).Infof("%s is not going to run due to dependent component disabled.", name)
			continue
		}

		if err := mgr.Add(ctrl); err != nil {
			klog.Error(err, "add controller to manager failed", "name", name)
			return err
		}
	}

	return nil
}
