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
	Upgrade() (*release.Release, error)
	Uninstall() (*release.UninstallReleaseResponse, error)
	Status(release string) ([]model.ClusterComponentResStatus, error)
}

func installChart(client *helm.Client, releaseName, chartName, chartVersion string, values map[string]interface{}) (*release.Release, error) {
	err := preInstallChart(client, releaseName)
	if err != nil {
		return nil, err
	}

	//m, err := MergeValueMap(values)
	//if err != nil {
	//	return nil, err
	//}
	// logger.Log.Infof("start install tool %s with chartName: %s, chartVersion: %s", tool.Name, chartName, chartVersion)
	release, err := client.Install(releaseName, chartName, chartVersion, values)
	if err != nil {
		return nil, err
	}
	// logger.Log.Infof("install tool %s successful", tool.Name)
	return release, nil
}

func upgradeChart(client *helm.Client, releaseName, chartName, chartVersion string, values map[string]interface{}) (*release.Release, error) {
	//m, err := MergeValueMap(values)
	//if err != nil {
	//	return nil, err
	//}
	klog.V(4).Infof("start upgrade tool %s with chartName: %s, chartVersion: %s", releaseName, chartName, chartVersion)
	rel, err := client.Upgrade(releaseName, chartName, chartVersion, values)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("upgrade tool %s successful", releaseName)
	return rel, nil
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

func uninstall(client *helm.Client, releaseName, ingressName, ingressVersion string) (*release.UninstallReleaseResponse, error) {
	rs, err := client.List()
	if err != nil {
		return nil, err
	}
	for _, r := range rs {
		if r.Name == releaseName {
			rel, err := client.Uninstall(releaseName)
			if err != nil {
				return nil, err
			}
			return rel, nil
		}
	}
	klog.V(4).Infof("uninstall component %s  successful", releaseName)

	// todo 删除ingress
	return nil, nil
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

func mergeMaps(map1 map[string]interface{}, map2 map[string]interface{}) map[string]interface{} {
	mergedMap := make(map[string]interface{})

	for k, v1 := range map1 {
		if v2, ok := map2[k]; ok {
			if subMap1, ok := v1.(map[string]interface{}); ok {
				if subMap2, ok := v2.(map[string]interface{}); ok {
					mergedMap[k] = mergeMaps(subMap1, subMap2)
				} else {
					mergedMap[k] = v1
				}
			} else {
				mergedMap[k] = v2
			}
		} else {
			mergedMap[k] = v1
		}
	}

	for k, v2 := range map2 {
		if _, ok := map1[k]; !ok {
			mergedMap[k] = v2
		}
	}

	return mergedMap
}

func getReleaseStatus(client *helm.Client, releaseName string) ([]model.ClusterComponentResStatus, error) {
	return client.Status(releaseName)
}
