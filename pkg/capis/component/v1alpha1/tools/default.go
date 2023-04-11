package tools

import (
	model "captain/pkg/models/component"
	"captain/pkg/simple/client/helm"

	"helm.sh/helm/v3/pkg/release"
)

type DefaultTool struct {
	client           *helm.Client
	clusterComponent *model.ClusterComponent

	release string
	chart   string
	version string
	values  map[string]interface{}
}

func (p DefaultTool) setDefaultValue() {
	p.values = mergeMaps(p.clusterComponent.DefultValues, p.clusterComponent.Parameters)
}

func NewDefaultTool(client *helm.Client, clusterComponent *model.ClusterComponent) (*DefaultTool, error) {
	p := &DefaultTool{
		client:           client,
		clusterComponent: clusterComponent,

		release: clusterComponent.ReleaseName,
		chart:   clusterComponent.ChartName,
		version: clusterComponent.ChartVersion,
		values:  clusterComponent.Parameters,
	}
	return p, nil
}

func (p *DefaultTool) Install() (*release.Release, error) {
	p.setDefaultValue()
	release, err := installChart(p.client, p.release, p.chart, p.version, p.values)
	if err != nil {
		return nil, err
	}

	return release, err
}

func (p *DefaultTool) Upgrade() (*release.Release, error) {
	p.setDefaultValue()
	rel, err := upgradeChart(p.client, p.release, p.chart, p.version, p.values)
	return rel, err
}

func (p *DefaultTool) Uninstall() (*release.UninstallReleaseResponse, error) {

	//需要kube client同样 还缺少namespace信息
	//创建ingress之后 需要删除ingress
	return uninstall(p.client, nil, p.release, "", "")
}

func (p *DefaultTool) Status(release string) ([]model.ClusterComponentResStatus, error) {
	return getReleaseStatus(p.client, release)
}
