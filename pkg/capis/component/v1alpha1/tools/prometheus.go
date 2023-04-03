package tools

import (
	model "captain/pkg/models/component"
	"captain/pkg/simple/client/helm"
	"fmt"

	"helm.sh/helm/v3/pkg/release"
)

type Prometheus struct {
	client           *helm.Client
	clusterComponent *model.ClusterComponent

	release string
	chart   string
	version string
	values  map[string]interface{}
}

func NewPrometheus(client *helm.Client, clusterComponent *model.ClusterComponent) (*Prometheus, error) {
	p := &Prometheus{
		client:           client,
		clusterComponent: clusterComponent,

		release: clusterComponent.ReleaseName,
		chart:   clusterComponent.ChartName,
		version: clusterComponent.ChartVersion,
	}
	return p, nil
}

func (p Prometheus) setDefaultValue(clusterComponent *model.ClusterComponent, isInstall bool) {

	values := map[string]interface{}{}
	//根据不同版本prometheus填充
	switch clusterComponent.ChartVersion {
	case "15.0.1", "15.10.1":
		values = p.valuse1501Binding(isInstall)

	}
	if isInstall {
		//安装，存储初始化
		if _, ok := values["server.retention"]; ok {
			values["server.retention"] = fmt.Sprintf("%vd", values["server.retention"])
		}
		if _, ok := values["server.persistentVolume.size"]; ok {
			values["server.persistentVolume.size"] = fmt.Sprintf("%vGi", values["server.persistentVolume.size"])
		}
		if va, ok := values["server.persistentVolume.enabled"]; ok {
			if hasPers, _ := va.(bool); !hasPers {
				delete(values, "server.nodeSelector.kubernetes\\.io/hostname")
			}
		}
	}
	p.values = values

}

func (p *Prometheus) Install() (*release.Release, error) {
	p.setDefaultValue(p.clusterComponent, true)
	release, err := installChart(p.client, p.release, p.chart, p.version, p.values)
	if err != nil {
		return nil, err
	}

	// TODO create ingress
	/*ingressItem := &Ingress{
		name:    constant.DefaultPrometheusIngressName,
		url:     constant.DefaultPrometheusIngress,
		service: constant.DefaultPrometheusServiceName,
		port:    80,
		version: p.Cluster.Version,
	}
	if err := createRoute(p.Cluster.Namespace, ingressItem, p.Cluster.KubeClient); err != nil {
		return err
	}*/
	/*
		if err := waitForRunning(p.Cluster.Namespace, constant.DefaultPrometheusDeploymentName, 1, p.Cluster.KubeClient); err != nil {
			return err
		}*/
	return release, err
}

func (p *Prometheus) Upgrade() error {
	p.setDefaultValue(p.clusterComponent, false)
	err := upgradeChart(p.client, p.release, p.chart, p.version, p.values)
	return err
}

func (p *Prometheus) Uninstall() (*release.UninstallReleaseResponse, error) {

	//需要kube client同样 还缺少namespace信息
	//创建ingress之后 需要删除ingress
	return uninstall(p.client, p.release, "", "")
}

func (p *Prometheus) Status(release string) ([]model.ClusterComponentResStatus, error) {
	return getReleaseStatus(p.client, release)
}

func (p Prometheus) valuse1501Binding(isInstall bool) map[string]interface{} {
	values := map[string]interface{}{}
	if len(p.clusterComponent.Parameters) != 0 {
		values = p.clusterComponent.Parameters
	}
	if !isInstall {
		//TODO
		//升级时  需要删除promtheus-kube-state-metrics  否则会触发一系列告警
		//kube client delete ns prometheus-kube-state-metrics
		//klog.V(4).Infof("delete deployment prometheus-kube-state-metrics from %s failed, err: %v", namespace, err)

	}
	values["alertmanager.enabled"] = false
	values["pushgateway.enabled"] = false
	values["configmapReload.prometheus.enabled"] = true
	values["nodeExporter.enabled"] = true
	values["server.enabled"] = true
	values["server.service.type"] = "NodePort"

	//image和tag的控制 需要在clusterComponent.Parameters中灵活制定
	return values
}
