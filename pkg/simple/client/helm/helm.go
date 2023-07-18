package helm

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	model "captain/pkg/models/component"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
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
	"k8s.io/klog"
)

func nolog(format string, v ...interface{}) {}

type Client struct {
	actionConfig *action.Configuration
	namespace    string
	settings     *cli.EnvSettings
	options      *Options
}

func NewClient(kubeConfig []byte, namespace string, options *Options) (*Client, error) {
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
		options:      options,
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
	err = actionConfig.Init(cf, namespace, "configmap", nolog)
	return actionConfig, err
}

func (c Client) Install(releaseName, chartName, chartVersion string, values map[string]interface{}) (*release.Release, error) {
	if err := c.updateRepo(chartName, ""); err != nil {
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
	client.Wait = true
	client.Timeout = 3 * time.Minute
	release, err := client.Run(releaseName)
	if err != nil {
		return release, fmt.Errorf("uninstall tool %s failed: %v", releaseName, err)
	}
	return release, nil
}

func (c Client) List() ([]*release.Release, error) {
	client := action.NewList(c.actionConfig)
	client.All = true
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
	resources, err := c.actionConfig.KubeClient.Build(bytes.NewBufferString(rel.Manifest), false)
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
	if err := c.updateRepo(chartName, ""); err != nil {
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

func (c Client) addRepo(arch string) error {
	settings := GetSettings()
	repoFile := settings.RepositoryConfig
	if err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}

	fileLock := flock.New(strings.Replace(repoFile, filepath.Ext(repoFile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer func() {
			if err := fileLock.Unlock(); err != nil {
				klog.Errorf("addRepo fileLock.Unlock failed, error: %s", err.Error())
			}
		}()
	}
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	// get repo
	e := c.getComponentRepo("", "")
	r, err := repo.NewChartRepository(e, getter.All(settings))
	if err != nil {
		return err
	}
	r.CachePath = settings.RepositoryCache
	if _, err := r.DownloadIndexFile(); err != nil {
		return errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", e.URL)
	}

	f.Update(e)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return err
	}
	return nil
}

func updateCharts() error {
	klog.V(4).Infoln("Hang tight while we grab the latest from your chart repositories...")
	settings := GetSettings()
	repoFile := settings.RepositoryConfig
	repoCache := settings.RepositoryCache
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("load file of repo %s failed: %v", repoFile, err)
	}
	var rps []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			return fmt.Errorf("get new chart repository failed, err: %v", err.Error())
		}
		if repoCache != "" {
			r.CachePath = repoCache
		}
		rps = append(rps, r)
	}

	var wg sync.WaitGroup
	for _, re := range rps {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				klog.V(4).Infof("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
			} else {
				klog.V(4).Infof("...Successfully got an update from the %q chart repository\n", re.Config.Name)
			}
		}(re)
	}
	wg.Wait()
	klog.V(4).Infof("Update Complete. ⎈ Happy Helming!⎈ ")
	return nil
}

// 每次安装或升级组件的时候执行
// 每次启用或升级的时候执行，存在 nexus 则不采取操作？
func (c Client) updateRepo(component, arch string) error {
	/*repos, err := ListRepo()
	if err != nil {
		klog.V(4).Infof("list repo failed: %v, start reading from db repo", err)
	}
	flag := false
	for _, r := range repos {
		if r.Name == "nexus" {
			klog.V(4).Infof("my nexus addr is %s", r.URL)
			flag = true
		}
	}
	if !flag {
	*/
	if err := c.addRepo(arch); err != nil {
		return err
	}
	if err := updateCharts(); err != nil {
		return err
	}
	// }
	return nil
}

func (c Client) getComponentRepo(component, arch string) *repo.Entry {
	// TODO switch component case: return repoUrl and Name
	return &repo.Entry{
		Name:                  c.options.Name,
		URL:                   c.options.URL,
		Username:              c.options.Username,
		Password:              c.options.Password,
		InsecureSkipTLSverify: true,
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

func ListRepo() ([]*repo.Entry, error) {
	settings := GetSettings()
	var repos []*repo.Entry
	f, err := repo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		return repos, err
	}
	return f.Repositories, nil
}
