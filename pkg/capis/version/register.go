package version

import (
	"captain/pkg/server/runtime"
	"captain/pkg/version"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

func AddToContainer(container *restful.Container, k8sDiscovery discovery.DiscoveryInterface, k8sclient kubernetes.Interface) error {
	webservice := runtime.NewWebService(schema.GroupVersion{})

	webservice.Route(webservice.GET("/version").
		To(func(request *restful.Request, response *restful.Response) {
			captainVersion := version.Get()

			if k8sDiscovery != nil {
				k8sVersion, err := k8sDiscovery.ServerVersion()
				if err == nil {
					captainVersion.Kubernetes = k8sVersion
				} else {
					klog.Errorf("Failed to get kubernetes version, error %v", err)
				}
			}

			response.WriteAsJson(captainVersion)
		})).
		Doc("Captain version")

	container.Add(webservice)

	return nil
}