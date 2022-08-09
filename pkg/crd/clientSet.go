package crd

import (
	"captain/pkg/client/clientset/versioned"
	"captain/pkg/crd/v1beta1"

	"k8s.io/client-go/rest"
)

//　all crd interface set
// to avoid interface duplicate name, you need to implement crd such as Getxxx
// for example GaiaCluster crd:
// 		call as:  CrdClientSet.GaiaClusters.Create(...)...
type CrdInterface interface {
	V1beta1() v1beta1.V1beta1Interface
	Versioned() versioned.Interface
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
	versioned *versioned.Clientset
}

func (c *CrdController) V1beta1() v1beta1.V1beta1Interface {
	return c.v1beta1
}

func (c *CrdController) Versioned() versioned.Interface {
	return c.versioned
}

func NewForConfig(c *rest.Config) (CrdInterface, error) {
	configShallowCopy := *c

	if configShallowCopy.UserAgent == "" {
		configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	controller := &CrdController{}

	// versioned
	versionedResource, err := versioned.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	controller.versioned = versionedResource

	// v1beta1
	v1beta1Resource, err := v1beta1.NewV1beta1Config(c, versionedResource)
	if err != nil {
		return nil, err
	}
	controller.v1beta1 = v1beta1Resource

	// todo list
	return controller, nil
}
