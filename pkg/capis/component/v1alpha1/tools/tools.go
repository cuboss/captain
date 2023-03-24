package tools

import (
	"encoding/json"
	"fmt"

	model "captain/pkg/models/component"
	"captain/pkg/simple/client/helm"

	"helm.sh/helm/v3/pkg/release"
	"k8s.io/helm/pkg/strvals"
)

type Interface interface {
	Install() (*release.Release, error)
	Upgrade() error
	Uninstall() error
	Status()
}

func installChart(client *helm.Client, clusterComponent *model.ClusterComponent) (*release.Release, error) {
	err := preInstallChart(client, clusterComponent.ReleaseName)
	if err != nil {
		return nil, err
	}
	valueMap := map[string]interface{}{}
	_ = json.Unmarshal([]byte(clusterComponent.Values), &valueMap)
	m, err := MergeValueMap(valueMap)
	if err != nil {
		return nil, err
	}
	// logger.Log.Infof("start install tool %s with chartName: %s, chartVersion: %s", tool.Name, chartName, chartVersion)
	release, err := client.Install(clusterComponent.ComponentName, clusterComponent.ClusterName, clusterComponent.ComponentVersion, clusterComponent.Namespace, m)
	if err != nil {
		return nil, err
	}
	// logger.Log.Infof("install tool %s successful", tool.Name)
	return release, nil
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
