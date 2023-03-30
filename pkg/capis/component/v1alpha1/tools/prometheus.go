package tools

import (
	model "captain/pkg/models/component"
	"captain/pkg/simple/client/helm"

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
	p.values = clusterComponent.Parameters
	// TODO
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
	return nil
}

func (p *Prometheus) Uninstall() error {
	return nil
}

func (p *Prometheus) Status(release string) ([]model.ClusterComponentResStatus, error) {
	return getReleaseStatus(p.client, release)
}
