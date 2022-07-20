package version

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/klog"

	"captain/pkg/apiserver/runtime"
	"captain/pkg/version"
)

func AddToContainer(container *restful.Container, k8sDiscovery discovery.DiscoveryInterface) error {
	webservice := runtime.NewWebService(schema.GroupVersion{})

	webservice.Route(webservice.GET("/version").
		To(func(request *restful.Request, response *restful.Response) {
			ksVersion := version.Get()

			if k8sDiscovery != nil {
				k8sVersion, err := k8sDiscovery.ServerVersion()
				if err == nil {
					ksVersion.Kubernetes = k8sVersion
				} else {
					klog.Errorf("Failed to get kubernetes version, error %v", err)
				}
			}

			response.WriteAsJson(ksVersion)
		})).
		Doc("Captain version")

	container.Add(webservice)

	return nil
}