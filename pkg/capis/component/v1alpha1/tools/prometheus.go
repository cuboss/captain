package tools

import (
	model "captain/pkg/models/component"
	"captain/pkg/simple/client/helm"
	"context"
	"fmt"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	DefaultPrometheusIngressName = "prometheus-ingress"
	DefaultPrometheusIngress     = "prometheus." + DefaultIngress
	DefaultIngress               = "apps.com"
	DefaultPrometheusServiceName = "prometheus-server"

	DefaultPrometheusDeploymentName = "prometheus-server"
)

type Prometheus struct {
	client           *helm.Client
	clusterComponent *model.ClusterComponent
	kubeClient       *kubernetes.Clientset
	release          string
	chart            string
	version          string
	values           map[string]interface{}
}

func NewPrometheus(client *helm.Client, kubeClient *kubernetes.Clientset, clusterComponent *model.ClusterComponent) (*Prometheus, error) {
	p := &Prometheus{
		client:           client,
		kubeClient:       kubeClient,
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

	ingressItem := &Ingress{
		name:    DefaultPrometheusIngressName,
		url:     DefaultPrometheusIngress,
		service: DefaultPrometheusServiceName,
		//version 暂时未添加
		port: 80,
	}
	if err := createRoute(p.clusterComponent.Namespace, ingressItem, p.kubeClient); err != nil {
		return nil, err
	}

	if err := waitForRunning(p.clusterComponent.Namespace, DefaultPrometheusDeploymentName, 1, p.kubeClient); err != nil {
		return nil, err
	}
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
	return uninstall(p.client, p.kubeClient, p.release, DefaultPrometheusIngressName, p.clusterComponent.Namespace)
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
		get, _ := p.kubeClient.AppsV1().Deployments(p.clusterComponent.Namespace).Get(context.TODO(), "prometheus-kube-state-metrics", v1.GetOptions{})
		//NotFound error不影响
		if get.Name != "" {
			if err := p.kubeClient.AppsV1().Deployments(p.clusterComponent.Namespace).Delete(context.TODO(), "prometheus-kube-state-metrics", v1.DeleteOptions{}); err != nil {
				klog.Infof("delete deployment prometheus-kube-state-metrics from %s failed, err: %v", p.clusterComponent.Namespace, err)
			}
		}

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
