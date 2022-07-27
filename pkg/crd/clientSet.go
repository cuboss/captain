package crd

import (
	"captain/pkg/crd/v1beta1"
	gaia2 "captain/pkg/crd/v1beta1/gaia"
	"k8s.io/client-go/rest"
)

//　all crd interface set
// to avoid interface duplicate name, you need to implement crd such as Getxxx
// for example GaiaCluster crd:
// 		call as:  CrdClientSet.GaiaClusters.Create(...)...
type CrdInterface interface {
	V1beta1() v1beta1.V1beta1Interface
	// todo list
	// crd interface
}

// all crd client set
// crd client aggregate into one clientSet which has the same purpose
// reference GaiaCrdSet
type CrdController struct {
	v1beta1 *v1beta1.V1beta1Client
	// todo list
	// crd client
}

func (c *CrdController) V1beta1() v1beta1.V1beta1Interface {
	return c.v1beta1
}

func NewForConfig(c *rest.Config) (CrdInterface, error) {
	configShallowCopy := *c

	if configShallowCopy.UserAgent == "" {
		configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	controller := &CrdController{
	}
	// v1beta1
	v1beta1Resource, err := v1beta1.NewV1beta1Config(c)
	if err != nil {
		return nil, err
	}

	controller.v1beta1 = v1beta1Resource
    // todo list
	return controller, nil
}


func (c *CrdController) GaiaClusters(namespace string) gaia2.GaiaClusterClient {
	return c.v1beta1.GaiaCluster(namespace)
}

// ClusterSets 获取ClusterSet Client
func (c *CrdController) GaiaNodes(namespace string) gaia2.GaiaNodeClient {
	return c.v1beta1.GaiaNode(namespace)
}

// ClusterSets 获取ClusterSet Client
func (c *CrdController) GaiaSets(namespace string) gaia2.GaiaSetClient {
	return c.v1beta1.GaiaSet(namespace)
}
