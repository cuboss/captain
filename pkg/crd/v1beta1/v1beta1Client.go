package v1beta1

import (
	"captain/pkg/crd/v1beta1/gaia"
	"k8s.io/client-go/rest"
)
type V1beta1Interface interface {
	gaia.GaiaClusterGetter
	gaia.GaiaNodeGetter
	gaia.GaiaSetGetter

	// todo list
}

type V1beta1Client struct {
	GaiaCliet *gaia.GaiaCrdSet
	// todo list
}

func NewV1beta1Config(c *rest.Config) (*V1beta1Client, error) {
	configShallowCopy := *c

	if configShallowCopy.UserAgent == "" {
		configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	controller := &V1beta1Client{
	}
	// gaia
	gaiacrdSet, err := gaia.NewGaiaCrd(c)
	if err != nil {
		return nil, err
	}
	controller.GaiaCliet = gaiacrdSet
	// todo list

	return controller, nil
}

func (c *V1beta1Client) GaiaCluster(namespace string) gaia.GaiaClusterClient {
	return c.GaiaCliet.GaiaClusters(namespace)
}

func (c *V1beta1Client) GaiaNode(namespace string) gaia.GaiaNodeClient {
	return c.GaiaCliet.GaiaNodes(namespace)
}

func (c *V1beta1Client) GaiaSet(namespace string) gaia.GaiaSetClient {
	return c.GaiaCliet.GaiaSets(namespace)
}



