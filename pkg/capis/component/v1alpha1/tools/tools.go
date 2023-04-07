package tools

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"

	netv1 "k8s.io/api/networking/v1"
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

type Ingress struct {
	name    string
	url     string
	service string
	port    int
}

func createRoute(namespace string, ingressInfo *Ingress, kubeClient *kubernetes.Clientset) error {
	if err := preCreateRoute(namespace, ingressInfo.name, kubeClient); err != nil {
		return err
	}
	service, err := kubeClient.CoreV1().
		Services(namespace).
		Get(context.TODO(), ingressInfo.service, metav1.GetOptions{})
	if err != nil {
		return err
	}

	ingressInfo.service = service.Name

	ingress := newNetwork(namespace, ingressInfo)
	if _, err = kubeClient.NetworkingV1().Ingresses(namespace).Create(context.TODO(), ingress, metav1.CreateOptions{}); err != nil {
		return err
	}

	klog.Infof("create route %s successful", ingressInfo.name)
	return nil
}

func preCreateRoute(namespace string, ingressName string, kubeClient *kubernetes.Clientset) error {

	ingress, _ := kubeClient.NetworkingV1().Ingresses(namespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if ingress.Name != "" {
		if err := kubeClient.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), ingressName, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	klog.Infof("operation before create route %s successful", ingressName)
	return nil
}

func newNetwork(namespace string, ingressInfo *Ingress) *netv1.Ingress {
	pathType := netv1.PathTypePrefix
	ingress := netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressInfo.name,
			Namespace: namespace,
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{
				{
					Host: ingressInfo.url,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: ingressInfo.service,
											Port: netv1.ServiceBackendPort{
												Number: int32(ingressInfo.port),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return &ingress
}

func waitForRunning(namespace string, deploymentName string, minReplicas int32, kubeClient *kubernetes.Clientset) error {
	klog.Infof("installation and configuration successful, now waiting for %s running", deploymentName)
	kubeClient.CoreV1()
	err := wait.Poll(5*time.Second, 10*time.Minute, func() (done bool, err error) {
		d, err := kubeClient.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			return true, err
		}
		if d.Status.ReadyReplicas > minReplicas-1 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
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

func upgradeChart(client *helm.Client, releaseName, chartName, chartVersion string, values map[string]interface{}) (*release.Release, error) {
	m, err := MergeValueMap(values)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("start upgrade tool %s with chartName: %s, chartVersion: %s", releaseName, chartName, chartVersion)
	rel, err := client.Upgrade(releaseName, chartName, chartVersion, m)
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

func uninstall(client *helm.Client, kubeClient *kubernetes.Clientset, releaseName, ingressName, namespace string) (*release.UninstallReleaseResponse, error) {
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

	//get, _ := kubeClient.NetworkingV1().Ingresses(namespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	//NotFound error不影响
	if err := kubeClient.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), ingressName, metav1.DeleteOptions{}); err != nil {
		klog.Errorf("uninstall tool %s of namespace %s failed, err: %v", releaseName, namespace, err)
	}

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

func getReleaseStatus(client *helm.Client, releaseName string) ([]model.ClusterComponentResStatus, error) {
	return client.Status(releaseName)
}
