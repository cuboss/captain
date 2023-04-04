package helm

import (
	"bytes"
	"context"
	"fmt"
	"log"

	model "captain/pkg/models/component"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	actionConfig *action.Configuration
	namespace    string
	settings     *cli.EnvSettings
}

func NewClient(kubeConfig []byte, namespace string) (*Client, error) {
	if namespace == "" {
		namespace = model.DefaultNamespace
	}
	actionConfig, err := initActionConfig(kubeConfig, namespace)
	if err != nil {
		return nil, fmt.Errorf("initActionConfig error: %s", err.Error())
	}
	// TODO init client
	client := Client{
		actionConfig: actionConfig,
		settings:     GetSettings(),
		namespace:    namespace,
	}

	return &client, nil
}

func initActionConfig(kubeconfig []byte, namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	cf := genericclioptions.NewConfigFlags(true)
	kconfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	cf.WrapConfigFn = func(config *rest.Config) *rest.Config {
		return kconfig
	}
	err = actionConfig.Init(cf, namespace, "configmap", log.Printf)
	return actionConfig, err
}

func (c Client) Install(releaseName, chartName, chartVersion string, values map[string]interface{}) (*release.Release, error) {
	if err := updateRepo(chartName); err != nil {
		return nil, err
	}
	client := action.NewInstall(c.actionConfig)
	client.ReleaseName = releaseName
	client.Namespace = c.namespace
	client.ChartPathOptions.InsecureSkipTLSverify = true
	if len(chartVersion) != 0 {
		client.ChartPathOptions.Version = chartVersion
	}
	p, err := client.ChartPathOptions.LocateChart(chartName, c.settings)
	if err != nil {
		return nil, fmt.Errorf("locate chart %s failed: %v", chartName, err)
	}
	ct, err := loader.Load(p)
	if err != nil {
		return nil, fmt.Errorf("load chart %s failed: %v", chartName, err)
	}
	release, err := client.Run(ct, values)
	if err != nil {
		return release, fmt.Errorf("install tool %s with chart %s failed: %v", releaseName, chartName, err)
	}
	return release, nil
}

func (c Client) Uninstall(releaseName string) (*release.UninstallReleaseResponse, error) {
	client := action.NewUninstall(c.actionConfig)

	release, err := client.Run(releaseName)
	if err != nil {
		return release, fmt.Errorf("uninstall tool %s failed: %v", releaseName, err)
	}
	return release, nil
}

func (c Client) List() ([]*release.Release, error) {
	client := action.NewList(c.actionConfig)
	release, err := client.Run()
	if err != nil {
		return release, fmt.Errorf("list chart failed: %v", err)
	}
	return release, nil
}

func (c Client) Status(releaseName string) ([]model.ClusterComponentResStatus, error) {
	ress := make([]model.ClusterComponentResStatus, 0)

	// get release and list resource
	client := action.NewStatus(c.actionConfig)
	rel, err := client.Run(releaseName)
	if err != nil {
		return nil, err
	}
	resources, err := c.actionConfig.KubeClient.Build(bytes.NewBufferString(rel.Manifest), true)
	if err != nil {
		return nil, fmt.Errorf("unable to build kubernetes objects from release manifest, err: %v", err)
	}
	// init k8s cli and for-check resource
	config, err := c.actionConfig.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	cli, err := kubernetes.NewForConfig(config)
	message := ""
	checker := kube.NewReadyChecker(cli, func(format string, a ...interface{}) {
		message = fmt.Sprintf(format, a...)
	}, kube.PausedAsReady(true))
	for _, v := range resources {
		ready, err := checker.IsReady(context.Background(), v)
		if err != nil {
			return nil, err
		}
		ress = append(ress, model.ClusterComponentResStatus{
			Kind:      v.Mapping.GroupVersionKind.Kind,
			Name:      v.Name,
			Namespace: v.Namespace,
			Ready:     ready,
			Message:   message,
		})
	}
	return ress, nil
}

func (c Client) Upgrade(releaseName, chartName, chartVersion string, values map[string]interface{}) (*release.Release, error) {

	if err := updateRepo(chartName); err != nil {
		return nil, err
	}
	client := action.NewUpgrade(c.actionConfig)
	//执行升级
	client.DryRun = false
	client.ChartPathOptions.InsecureSkipTLSverify = true
	client.ChartPathOptions.Version = chartVersion
	client.Namespace = c.namespace
	p, err := client.ChartPathOptions.LocateChart(chartName, c.settings)
	if err != nil {
		return nil, fmt.Errorf("locate chart %s failed: %v", chartName, err)
	}
	//loader执行load方法
	ct, err := loader.Load(p)
	if err != nil {
		return nil, fmt.Errorf("load chart %s failed: %v", chartName, err)
	}

	release, err := client.Run(releaseName, ct, values)
	if err != nil {
		return release, fmt.Errorf("upgrade tool %s with chart %s failed: %v", releaseName, chartName, err)
	}
	return release, nil
}

// 每次安装或升级组件的时候执行
func updateRepo(component string) error {
	helmRepo, err := repo.NewChartRepository(getComponentRepo(component), getter.All(GetSettings()))
	if err != nil {
		return fmt.Errorf("Failed to create Helm repository: %v\n", err)
	}
	// 更新Helm仓库索引
	if _, err := helmRepo.DownloadIndexFile(); err != nil {
		return fmt.Errorf("Failed to update Helm repository index: %v\n", err)
	}
	return nil
}

func getComponentRepo(component string) *repo.Entry {
	// TODO switch component case: return repoUrl and Name
	// 创建一个Helm仓库对象
	return &repo.Entry{
		Name: "mirantis",
		// 设置Helm仓库的地址
		URL: "https://charts.mirantis.com/",
	}
}

func GetSettings() *cli.EnvSettings {
	return &cli.EnvSettings{
		PluginsDirectory: helmpath.DataPath("plugins"),
		RegistryConfig:   helmpath.ConfigPath("registry.json"),
		RepositoryConfig: helmpath.ConfigPath("repositories.yaml"),
		RepositoryCache:  helmpath.CachePath("repository"),
	}
}
