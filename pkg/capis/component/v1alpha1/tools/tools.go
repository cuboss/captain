package tools

import (
	"fmt"
	"k8s.io/klog"

	model "captain/pkg/models/component"
	"captain/pkg/simple/client/helm"

	"helm.sh/helm/v3/pkg/release"
	"k8s.io/helm/pkg/strvals"
)

type Interface interface {
	Install() (*release.Release, error)
	Upgrade() error
	Uninstall() error
	Status(release string) ([]model.ClusterComponentResStatus, error)
}

func installChart(client *helm.Client, releaseName, chartName, chartVersion string, values map[string]interface{}) (*release.Release, error) {
	err := preInstallChart(client, releaseName)
	if err != nil {
		return nil, err
	}

	m, err := MergeValueMap(values)
	if err != nil {
		return nil, err
	}
	// logger.Log.Infof("start install tool %s with chartName: %s, chartVersion: %s", tool.Name, chartName, chartVersion)
	release, err := client.Install(releaseName, chartName, chartVersion, m)
	if err != nil {
		return nil, err
	}
	// logger.Log.Infof("install tool %s successful", tool.Name)
	return release, nil
}

func upgradeChart(client *helm.Client, releaseName, chartName, chartVersion string, values map[string]interface{}) error {
	m, err := MergeValueMap(values)
	if err != nil {
		return err
	}
	klog.V(4).Infof("start upgrade tool %s with chartName: %s, chartVersion: %s", releaseName, chartName, chartVersion)
	_, err = client.UpGrade(releaseName, chartName, chartVersion, m)
	if err != nil {
		return err
	}
	klog.V(4).Infof("upgrade tool %s successful", releaseName)
	return nil
}

func preInstallChart(client *helm.Client, releaseName string) error {
	rs, err := client.List()
	if err != nil {
		return err
	}
	for _, r := range rs {
		if r.Name == releaseName {
			// LOG logger.Log.Infof("uninstall %s before installation", tool.Name)
			_, err := client.Uninstall(releaseName)
			if err != nil {
				return err
			}
		}
	}
	// logger.Log.Infof("uninstall %s before installation successful", tool.Name)
	return nil
}

func uninstall(client *helm.Client, releaseName, ingressName, ingressVersion string) error {

	rs, err := client.List()
	if err != nil {
		return err
	}
	for _, r := range rs {
		if r.Name == releaseName {
			_, err := client.Uninstall(releaseName)
			if err != nil {
				return err
			}
		}
	}

	//删除ingress
	//todo
	//缺少client和namespace信息

	//logger.Log.Infof("uninstall tool %s of namespace %s successful", tool.Name, namespace)
	return nil
}

func MergeValueMap(source map[string]interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	var valueStrings []string
	for k, v := range source {
		str := fmt.Sprintf("%s=%v", k, v)
		valueStrings = append(valueStrings, str)
	}
	for _, str := range valueStrings {
		err := strvals.ParseInto(str, result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func getReleaseStatus(client *helm.Client, releaseName string) ([]model.ClusterComponentResStatus, error) {
	return client.Status(releaseName)
}
