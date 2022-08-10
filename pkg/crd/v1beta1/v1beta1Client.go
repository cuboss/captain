package v1beta1

import (
	"k8s.io/client-go/rest"

	"captain/pkg/client/clientset/versioned"
	clusterv1alpha1 "captain/pkg/client/clientset/versioned/typed/cluster/v1alpha1"
	"captain/pkg/crd/v1beta1/gaia"
)

type V1beta1Interface interface {
	gaia.GaiaClusterGetter
	gaia.GaiaNodeGetter
	gaia.GaiaSetGetter

	clusterv1alpha1.ClustersGetter
	// todo list
}

type V1beta1Client struct {
	GaiaCliet *gaia.GaiaCrdSet
	// cluster
	Versioned *versioned.Clientset
	// todo list
}

func NewV1beta1Config(c *rest.Config, versioned *versioned.Clientset) (*V1beta1Client, error) {
	configShallowCopy := *c

	if configShallowCopy.UserAgent == "" {
		configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	controller := &V1beta1Client{}
	// gaia
	gaiacrdSet, err := gaia.NewGaiaCrd(c)
	if err != nil {
		return nil, err
	}
	controller.GaiaCliet = gaiacrdSet
	// versioned
	controller.Versioned = versioned
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

func (c *V1beta1Client) Clusters() clusterv1alpha1.ClusterInterface {
	return c.Versioned.ClusterV1alpha1().Clusters()
}
